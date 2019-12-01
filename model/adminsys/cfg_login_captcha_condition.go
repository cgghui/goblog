package adminsys

import "goblog/model/config"

// LoginCaptchaCondition 管理员登录时启用验证码的条件 cinfig: login_captcha_condition
type LoginCaptchaCondition struct {
	Password int `json:"pwd_errn"`
	Captcha  int `json:"captcha_errn"`
	admin    *Admins
}

// NewLoginCaptchaCondition 实例
func NewLoginCaptchaCondition(admin *Admins) *LoginCaptchaCondition {
	ret := LoginCaptchaCondition{}
	config.GetConfigField("admin", "login_captcha_condition").BindStruct(&ret)
	ret.admin = admin
	return &ret
}

// Check PasswordMax 和 CaptchaMax 只有一个为true 则返回true
func (l *LoginCaptchaCondition) Check() bool {
	return l.PasswordMax() || l.CaptchaMax()
}

// PasswordMax 如果密码错误次数大于条件设定 返回true
func (l *LoginCaptchaCondition) PasswordMax() bool {
	n, _ := l.admin.CounterGet(CounterPassword)
	return n > l.Password
}

// CaptchaMax 如果验证码错误次数大于条件设定 返回true
func (l *LoginCaptchaCondition) CaptchaMax() bool {
	n, _ := l.admin.CounterGet(CounterCaptcha)
	return n > l.Password
}
