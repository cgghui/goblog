package admin

import (
	"bytes"
	"encoding/json"
	"goblog/model/config"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"time"
)

// FilterCallAction 见下方
type FilterCallAction struct {
	Usable  bool   `json:"usable"`
	Type    string `json:"type"`
	Run     string `json:"run"`
	Timeout int64  `json:"timeout"`
}

// FilterCall 管理员登录时 外部调用的参数
// Before 登录开始之前
// After  登录成功之后
type FilterCall struct {
	Before   FilterCallAction `json:"before"`
	After    FilterCallAction `json:"after"`
	FormData map[string]interface{}
}

// NewFilterCall 实例化外部调用（登录时）
func NewFilterCall(form map[string]interface{}) *FilterCall {
	ret := FilterCall{}
	config.GetConfigField("admin", "login_filter_callback").BindStruct(&ret)
	ret.FormData = form
	return &ret
}

// FilterCallBefore 登录之前
func (f *FilterCall) FilterCallBefore() bool {
	if !f.Before.Usable {
		return true
	}
	ret, err := f.Before.callback(&f.FormData)
	if err != nil {
		panic(err)
	}
	return ret
}

// FilterCallAfter 登录之后
func (f *FilterCall) FilterCallAfter() bool {
	if !f.Before.Usable {
		return true
	}
	ret, err := f.After.callback(&f.FormData)
	if err != nil {
		panic(err)
	}
	return ret
}

func (f *FilterCallAction) callback(form *map[string]interface{}) (bool, error) {
	var (
		body []byte
		ret  *io.ReadCloser
		err  error
	)
	body, err = json.Marshal(form)
	if err != nil {
		return false, err
	}
	// URL 方法回调
	if f.Type == "url" {
		req, err := http.NewRequest("POST", f.Run, bytes.NewReader(body))
		if err != nil {
			return false, err
		}
		req.Header.Add("Content-Type", "application/json;charset=UTF-8")
		cli := &http.Client{
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					conn, err := net.DialTimeout(netw, addr, time.Duration(f.Timeout)*time.Second)
					if err != nil {
						return nil, err
					}
					return conn, nil
				},
			},
		}
		resp, err := cli.Do(req)
		if err != nil {
			return false, err
		}
		ret = &resp.Body
		defer resp.Body.Close()
	} else {
		cmd := exec.Command(f.Run, string(body))
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return false, err
		}
		ret = &stdout
		defer stdout.Close()
		if err := cmd.Start(); err != nil {
			return false, err
		}
		if err := cmd.Wait(); err != nil {
			return false, err
		}
	}
	result, err := ioutil.ReadAll(*ret)
	if err != nil {
		return false, err
	}
	return string(result) == "SUCCESS", nil
}
