package adminsys

import (
	"goblog/app"
	"goblog/model/config"

	"github.com/mojocn/base64Captcha"
)

// FormAdminLogin 登录表单
type FormAdminLogin struct {
	Username     string `form:"username"`
	Password     string `form:"password"`
	CaptchaCode  string `form:"captcha_code"`
	CaptchaToken string `form:"captcha_token"`
}

// GetAdmin 登录账号
func (f FormAdminLogin) GetAdmin() *Admins {
	admin := &Admins{}
	app.DBConn.Where("username = ?", f.Username).First(admin)
	return admin
}

// CheckCaptchaQuantity 验证验证码数量是否一致
func (f FormAdminLogin) CheckCaptchaQuantity() bool {
	conf := base64Captcha.ConfigCharacter{}
	config.GetConfigField("admin", "login_captcha_config").BindStruct(&conf)
	return conf.CaptchaLen == len(f.CaptchaCode)
}
