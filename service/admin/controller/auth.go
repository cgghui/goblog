package controller

import (
	"goblog/app"
	"goblog/model/adminsys"

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

	admin := &adminsys.Admins{}
	app.DBConn.Where("username = ?", username).First(admin)

	if !admin.Has() {
		app.Output(gin.H{"username": username}).DisplayJSON(ctx, app.StatusUserNotExist)
		return
	}

	admin.BuildKeyToRSA()

	data := gin.H{
		"locked":          true,
		"unlock_ttl":      86400,
		"pubkey":          "",
		"captcha_is_open": false,
		"captcha":         gin.H{},
	}

	if s := admin.CheckLocked(); s != 0 {
		data["locked"] = true
		data["unlock_ttl"] = s
		app.Output(data).DisplayJSON(ctx, app.StatusOK)
		return
	}

	data["pubkey"] = admin.PasswordCreateEncryptRSA()

	if admin.CaptchaLoginCheck() {
		img, token := admin.CaptchaLaod()
		data["captcha_is_open"] = true
		data["captcha"] = gin.H{"image": *img, "token": token}
	}

	app.Output(data).DisplayJSON(ctx, app.StatusOK)
}
