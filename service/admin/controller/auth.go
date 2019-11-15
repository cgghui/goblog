package controller

import (
	"fmt"
	"goblog/app"

	"github.com/gin-gonic/gin"
)

// Auth 授权
type Auth struct {
	*app.App
}

//Construct 构造方法
func (o *Auth) Construct(app *app.App) {
	app.GET("/auth/params", o.params)
	app.GET("/auth/params2", o.params2)
}

// AuthorizeInput 授权提交的内容
type AuthorizeInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (o *Auth) params(ctx *gin.Context) {
	app.Output().Assgin("a", "11111")
}

func (o *Auth) params2(ctx *gin.Context) {
	fmt.Println(app.Output().Get("a"))
}
