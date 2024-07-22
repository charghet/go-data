package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	uhttp "go-data/util/http"

	_ "github.com/mattn/go-sqlite3"
)

var dbInfo DatabaseInfo

type GetDataReq struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Str      bool   `json:"str"`
}

func GetDataServer(w http.ResponseWriter, req *http.Request) {
	var reqData GetDataReq
	err := uhttp.GetBodyEntity(req, &reqData)
	if err != nil {
		slog.Error("GetBodyJson error", "err", err)
		return
	}
	slog.Info("login", "userName", reqData.Name)
	content, err := dbInfo.GetContent(reqData.Name, reqData.Password)
	if err != nil {
		slog.Error("GetContent error", "err", err)
		return
	}
	var data any
	if reqData.Str && content != nil {
		data = string(content)
	} else {
		data = content
	}
	res := uhttp.Result{}.Ok(data)
	uhttp.SetResBody(w, res)
}

func main() {
	exitCode := 0
	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
		os.Exit(exitCode)
	}()

	help := flag.Bool("help", false, "Show help")

	host := flag.String("host", "localhost", "Host")
	port := flag.Int("port", 80, "Port")

	isUserList := flag.Bool("list", false, "Show user list")
	toAddUser := flag.Bool("adduser", false, "Add user to database")
	toDelUser := flag.Bool("deluser", false, "Delete user from database")
	userName := flag.String("user", "", "User name")
	password := flag.String("passwd", "", "User Password")

	// 解析命令行参数
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	// 连接数据库
	dbInfo = DatabaseInfo{
		DriverName:   "sqlite3",
		DatabaseName: "./data.db",
	}
	err := dbInfo.Open()
	if err != nil {
		slog.Error("Open database error", "err", err)
		exitCode = 1
		return
	}
	defer dbInfo.Close()

	// 显示用户列表
	if *isUserList {
		users, err := dbInfo.UserList()
		if err != nil {
			fmt.Println(err)
			exitCode = 1
			return
		}
		for _, user := range users {
			fmt.Println(user)
		}
		return
	}

	// 添加用户到数据库
	if *toAddUser {
		if *userName == "" || *password == "" {
			fmt.Println("user or passwd is empty!")
			exitCode = 1
		} else {
			err := dbInfo.AddUser(*userName, *password)
			if err != nil {
				fmt.Println(err)
				exitCode = 1
			} else {
				fmt.Printf("add user %s success", *userName)
			}
		}
		return
	}

	// 删除用户
	if *toDelUser {
		if *userName == "" {
			fmt.Println("user is empty!")
			exitCode = 1
		} else {
			err := dbInfo.DelUser(*userName)
			if err != nil {
				fmt.Println(err)
				exitCode = 1
			} else {
				fmt.Printf("del user %s success", *userName)
			}
			return
		}
	}

	// 启动HTTP服务
	fmt.Println("Start HTTP server")
	http.HandleFunc("/data", GetDataServer)
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", *host, *port), nil)
	if err != nil {
		fmt.Println(err)
		exitCode = 1
		return
	}
	fmt.Println("Server is running")
}
