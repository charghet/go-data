package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Ok(data interface{}) Result {
	return Result{
		Code: 200,
		Msg:  "ok",
		Data: data,
	}
}
func Fail(msg string, data interface{}) Result {
	if msg == "" {
		msg = "fail"
	}
	return Result{
		Code: 300,
		Msg:  msg,
		Data: data,
	}
}

func Error(msg string, data interface{}) Result {
	if msg == "" {
		msg = "error"
	}
	return Result{
		Code: 500,
		Msg:  msg,
		Data: data,
	}
}

func GetBodyJson(req *http.Request) (map[string]interface{}, error) {
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	var reqJson map[string]interface{}
	err = json.Unmarshal(reqBytes, &reqJson)
	if err != nil {
		return nil, err
	}
	return reqJson, nil
}

func GetBodyEntity(req *http.Request, entity interface{}) error {
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(reqBytes, &entity)
	if err != nil {
		return err
	}
	return nil
}

func SetResBody(w http.ResponseWriter, body interface{}) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	fmt.Fprint(w, string(bodyBytes))
	w.Header().Set("Content-Type", "application/json")
	return nil
}
