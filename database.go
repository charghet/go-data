package main

import (
	"database/sql"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

type DatabaseInfo struct {
	DriverName   string
	DatabaseName string
	db           *sql.DB
}

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (dbInfo *DatabaseInfo) Open() (err error) {
	// 打开数据库
	db, err := sql.Open(dbInfo.DriverName, dbInfo.DatabaseName)
	if err != nil {
		return err
	}
	dbInfo.db = db

	// 创建表
	createTableQuery := `CREATE TABLE IF NOT EXISTS data (
		"user_name" TEXT PRIMARY KEY, -- 用户名
		"password" TEXT, -- 密码
		"content" BLOB, -- 内容
		"last_login_time" TEXT, -- 最后登录时间
		"last_login_ip" TEXT -- 最后登录IP
	)`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return err
	}
	return nil
}

func (dbInfo *DatabaseInfo) Close() {
	// 关闭数据库
	if dbInfo.db != nil {
		dbInfo.db.Close()
	}
}

// AddUser 向数据库中添加用户
// 参数user为待添加的用户名
// 参数password为待添加的密码，使用bcrypt进行加密处理
// 若添加成功，则返回nil，否则抛出panic异常
func (dbInfo *DatabaseInfo) AddUser(user string, password string) (err error) {
	// 使用bcrypt加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 添加用户
	insertQuery := `INSERT INTO data (user_name, password) VALUES (?, ?)`
	_, err = dbInfo.db.Exec(insertQuery, user, string(hash))
	if err != nil {
		return err
	}
	return nil
}

func (dbInfo *DatabaseInfo) DelUser(user string) (err error) {
	// 删除用户
	deleteQuery := `DELETE FROM data WHERE user_name = ?`
	_, err = dbInfo.db.Exec(deleteQuery, user)
	if err != nil {
		return err
	}
	return nil
}

func (dbInfo *DatabaseInfo) UserList() (users []string, err error) {
	// 获取用户列表
	getUserInfoQuery := `SELECT user_name FROM data`
	rows, err := dbInfo.db.Query(getUserInfoQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user string
		err = rows.Scan(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// ResetPassword 函数用于重置数据库中某个用户的密码
// 参数user为需要重置密码的用户名
// 参数password为新的密码
// 该函数使用bcrypt库对密码进行加密，并更新数据库中对应用户的密码字段
// 如果加密或更新数据库时发生错误，则抛出panic异常
func (dbInfo *DatabaseInfo) ResetPassword(user string, password string) {
	// 使用bcrypt加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	// 重置密码
	updateQuery := `UPDATE data SET password = ? WHERE user_name = ?`
	_, err = dbInfo.db.Exec(updateQuery, string(hash), user)
	if err != nil {
		panic(err)
	}
}

func (dbInfo *DatabaseInfo) GetContent(user string, password string) (content []byte, err error) {
	// 获取用户信息
	getUserInfoQuery := `SELECT password,content FROM data WHERE user_name = ?`
	rows, err := dbInfo.db.Query(getUserInfoQuery, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hash := ""
	if rows.Next() {
		err = rows.Scan(&hash, &content)
		if err != nil {
			return nil, err
		}
		// 验证密码
		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		if err != nil {
			// 密码错误
			return nil, nil
		}
		if content == nil {
			content = []byte{}
		}
	}
	return content, nil
}
func (dbInfo *DatabaseInfo) SetContent(user string, password string, content []byte) (success bool, err error) {
	// 获取用户信息
	getUserInfoQuery := `SELECT password FROM data WHERE user_name = ?`
	rows, err := dbInfo.db.Query(getUserInfoQuery, user)
	if err != nil {
		return false, err
	}

	defer rows.Close()
	hash := ""
	if rows.Next() {
		err = rows.Scan(&hash)
		if err != nil {
			return false, err
		}
		// 验证密码
		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		if err != nil {
			// 密码错误
			slog.Warn("password error")
			return false, nil
		}
		// 更新内容
		rows.Close()
		updateQuery := `UPDATE data SET content = ? WHERE user_name = ?`
		_, err = dbInfo.db.Exec(updateQuery, content, user)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}
