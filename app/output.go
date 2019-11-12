package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Output 输出到浏览器
type Output struct {
	gin.H
}

// NewOutput 实例
func initOutput(app *App) {
	app.Output = &Output{
		H: gin.H{},
	}
}

// Assgin 添加一个值
// 已存在的将被替换 不存在的将被加入
func (o *Output) Assgin(field string, value interface{}) *Output {
	o.H[field] = value
	return o
}

// Del 删除一个值
func (o *Output) Del(field string, value interface{}) *Output {
	delete(o.H, field)
	return o
}

// DisplayHTML 显示HTML
func (o *Output) DisplayHTML(ctx *gin.Context, name string, code ...int) {
	if len(code) == 0 {
		ctx.HTML(http.StatusOK, name, o.H)
	} else {
		ctx.HTML(code[0], name, o.H)
	}
}
