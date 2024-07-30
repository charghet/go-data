package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	uhttp "go-data/util/http"

	_ "github.com/mattn/go-sqlite3"
)

var dbInfo DatabaseInfo

type LoginReq struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type GetDataReq struct {
	LoginReq
}

type SetDataReq struct {
	LoginReq
	Content string `json:"content"`
}

func GetDataHandler(w http.ResponseWriter, req *http.Request) {
	var reqData GetDataReq
	err := uhttp.GetBodyEntity(req, &reqData)
	if err != nil {
		slog.Error("GetBodyJson", "err", err)
		return
	}
	slog.Info("login", "userName", reqData.UserName)
	content, errMsg, err := dbInfo.GetContent(reqData.UserName, reqData.Password)

	if errMsg != "" {
		slog.Warn("GetContent", "warn", errMsg)
		uhttp.SetResBody(w, uhttp.Fail(errMsg, nil))
		return
	} else if err != nil {
		slog.Error("GetContent", "err", err.Error())
		uhttp.SetResBody(w, uhttp.Error("", nil))
		return
	}
	var data any = nil
	if content != nil {
		data = content
	}
	uhttp.SetResBody(w, uhttp.Ok(data))
}

func SetDataHandler(w http.ResponseWriter, req *http.Request) {
	var reqData SetDataReq
	err := uhttp.GetBodyEntity(req, &reqData)
	if err != nil {
		slog.Error("SetDataJson", "err", err)
		uhttp.SetResBody(w, uhttp.Error("", nil))
		return
	}

	bytes, err := base64.StdEncoding.DecodeString(reqData.Content)
	if err != nil {
		slog.Error("SetData", "err", err)
		uhttp.SetResBody(w, uhttp.Fail("base64 decode error!", nil))
		return
	}

	success, errMsg, err := dbInfo.SetContent(reqData.UserName, reqData.Password, bytes)
	if success {
		slog.Info("SetData success")
		uhttp.SetResBody(w, uhttp.Ok(nil))
	} else if errMsg != "" {
		slog.Warn("SetData", "warn", errMsg)
		uhttp.SetResBody(w, uhttp.Fail(errMsg, nil))
	} else if err != nil {
		slog.Error("SetData", "err", err)
		uhttp.SetResBody(w, uhttp.Error("", nil))
	}
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
	port := flag.Int("port", 3200, "Port")

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
	http.HandleFunc("/getData", GetDataHandler)
	http.HandleFunc("/setData", SetDataHandler)
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", *host, *port), nil)
	if err != nil {
		fmt.Println(err)
		exitCode = 1
		return
	}
	fmt.Println("Server is running")
}
