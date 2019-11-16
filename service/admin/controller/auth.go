package controller

import (
	"goblog/app"

	"github.com/gin-gonic/gin"
)

// Auth 授权
type Auth struct {
	*app.App
}

//Construct 构造方法
func (a *Auth) Construct(app *app.App) {
	app.GET("/auth/status", a.status)
}

// AuthorizeInput 授权提交的内容
type AuthorizeInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (a *Auth) status(ctx *gin.Context) {

	outoput := app.Output()

	username, ok := ctx.GetQuery("username")
	if !ok || len(username) == 0 {
		outoput.Assgin("info", "请输入账号")
		outoput.DisplayJSON(ctx, app.StatusQueryInvalid)
		return
	}

}
