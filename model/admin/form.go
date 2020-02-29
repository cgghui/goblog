package admin

import (
	"goblog/model/config"

	"github.com/mojocn/base64Captcha"
)

// FormLogin 后台管理员登录操作
type FormLogin struct {
	Username string `form:"username"`  // 提交过来的账号
	Password string `form:"password"`  // 提交过来的密码
	CaptchaC string `form:"captcha_c"` // 提交过来的验证码，明文码
	CaptchaT string `form:"captcha_t"` // 提交过的来的验证码校验字符串
}

// GetAdmin 登录账号
func (f FormLogin) GetAdmin() *Admins {
	return GetByUsername(f.Username)
}

// CheckCaptchaQuantity 验证码码数是否一致
func (f FormLogin) CheckCaptchaQuantity() bool {
	conf := base64Captcha.ConfigCharacter{}
	config.Get("admin", "login_captcha_config").BindStruct(&conf)
	return conf.CaptchaLen == len(f.CaptchaC)
}
