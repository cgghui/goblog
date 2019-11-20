package controller

import (
	"goblog/app"
	"goblog/model"

	"github.com/gin-gonic/gin"
)

// Auth 授权
type Auth struct {
}

//Construct 构造方法
func (a *Auth) Construct(app *app.App) {
	app.GET("/auth/check", a.check)
}

func (a *Auth) check(ctx *gin.Context) {

	username := ctx.Query("username")
	if len(username) == 0 {
		app.Output(gin.H{"tip": "请输入账号"}).DisplayJSON(ctx, app.StatusQueryInvalid)
		return
	}

	admin := &model.Admins{}
	app.DBConn.Where("username = ?", username).First(admin)

	if !admin.Has() {
		app.Output(gin.H{"username": username}).DisplayJSON(ctx, app.StatusUserNotExist)
		return
	}

	admin.Init()

	data := gin.H{
		"locked":          true,
		"unlock_ttl":      86400,
		"pubkey":          "",
		"captcha_is_open": false,
		"captcha":         gin.H{},
	}
	if model.AdminLoginCaptchaCheck(admin) {
		data["captcha_is_open"] = true
		data["captcha"] = gin.H{
			"image": "",
			"token": "",
		}
	}

	app.Output(data).DisplayJSON(ctx, app.StatusOK)
}
