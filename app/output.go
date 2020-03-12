package app

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// OutputFMT 输出到浏览器
type OutputFMT struct {
	gin.H
}

// OutputJSON 输出到浏览器的JSON格式
type OutputJSON struct {
	Status    bool   `json:"status"`
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Timestamp int64  `json:"timestamp"`
	Data      gin.H  `json:"data"`
}

// Output 输出到浏览器的对象
func Output(data ...gin.H) *OutputFMT {
	if len(data) == 0 {
		return &OutputFMT{gin.H{}}
	}
	return &OutputFMT{data[0]}
}

// Assgin 添加一个值
// 已存在的将被替换 不存在的将被加入
func (o *OutputFMT) Assgin(field string, value interface{}) *OutputFMT {
	o.H[field] = value
	return o
}

// Get 取出一个值
func (o *OutputFMT) Get(field string) (val interface{}, ok bool) {
	val, ok = o.H[field]
	return
}

// Del 删除一个值
func (o *OutputFMT) Del(field string) *OutputFMT {
	delete(o.H, field)
	return o
}

// DisplayHTML 显示HTML
func (o *OutputFMT) DisplayHTML(ctx *gin.Context, name string, code ...int) {
	if len(code) == 0 {
		ctx.HTML(http.StatusOK, name, o.H)
	} else {
		ctx.HTML(code[0], name, o.H)
	}
}

// DisplayJSON 显示JSON
func (o *OutputFMT) DisplayJSON(ctx *gin.Context, code int, args ...interface{}) {
	statusCode := http.StatusOK
	ascii := true
	switch len(args) {
	case 1:
		{
			switch val := args[0].(type) {
			case int:
				{
					statusCode = val
				}
			case bool:
				{
					ascii = val
				}
			}
		}
	case 2:
		{
			if val, ok := args[0].(int); ok {
				statusCode = val
			}
			if val, ok := args[1].(bool); ok {
				ascii = val
			}
		}
	}
	status, msg := StatusRet(code)
	output := OutputJSON{
		Status:    status,
		Code:      code,
		Msg:       msg,
		Timestamp: time.Now().Unix(),
	}
	if len(o.H) != 0 {
		output.Data = o.H
	}
	if ascii {
		ctx.AsciiJSON(statusCode, output)
	} else {
		ctx.JSON(statusCode, output)
	}
}
