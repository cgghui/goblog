package adminsys

import (
	"goblog/model/config"

	"github.com/mojocn/base64Captcha"
)

// FormAdminLogin 登录表单
type FormAdminLogin struct {
	Username     string `form:"username"`
	Password     string `form:"password"`
	CaptchaCode  string `form:"captcha_c"`
	CaptchaToken string `form:"captcha_t"`
}

// GetAdmin 登录账号
func (f FormAdminLogin) GetAdmin() *Admins {
	return GetAdminByUsernme(f.Username)
}

// CheckCaptchaQuantity 验证验证码数量是否一致
func (f FormAdminLogin) CheckCaptchaQuantity() bool {
	conf := base64Captcha.ConfigCharacter{}
	config.GetConfigField("admin", "login_captcha_config").BindStruct(&conf)
	return conf.CaptchaLen == len(f.CaptchaCode)
}
