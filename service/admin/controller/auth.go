package controller

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"goblog/app"
	"goblog/model"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
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

	admin.BuildKeyToRSA()

	data := gin.H{
		"locked":          true,
		"unlock_ttl":      86400,
		"pubkey":          admin.GetPasswordEncryptRSA(),
		"captcha_is_open": false,
		"captcha":         gin.H{},
	}

	if model.AdminLoginCaptchaCheck(admin) {
		data["captcha_is_open"] = true
		// 解析配置
		conf := base64Captcha.ConfigCharacter{}
		model.GetConfigField("admin", "login_captcha_config").BindJSON(&conf)
		// 构建验证码数据
		cimg := base64Captcha.EngineCharCreate(conf)
		verifyCode, err := json.Marshal([]interface{}{
			[]byte(cimg.VerifyValue),
			time.Now().Add(180 * time.Second).Unix(),
		})
		if err != nil {
			panic(fmt.Sprintf("Error: create captcha fail-1 %v", err))
		}
		// 加密数据
		token, err := rsa.EncryptPKCS1v15(rand.Reader, admin.PubKey, verifyCode)
		if err != nil {
			panic(fmt.Sprintf("Error: create captcha fail-2 %v", err))
		}
		data["captcha"] = gin.H{
			"image": base64Captcha.CaptchaWriteToBase64Encoding(cimg),
			"token": token,
		}
	}

	app.Output(data).DisplayJSON(ctx, app.StatusOK)
}
