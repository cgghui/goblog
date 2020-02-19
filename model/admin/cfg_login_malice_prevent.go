package admin

import (
	"goblog/model/config"
	"time"
)

// LoginMalicePrevent 管理员登录 密码错误次数限制（防止恶意尝试错误的密码）cinfig: login_malice_prevent
type LoginMalicePrevent struct {
	Password int   `json:"pwd_errn"`
	LockTime int64 `json:"lock_time"`
	admin    *Admins
}

// NewLoginMalicePrevent 实例
func NewLoginMalicePrevent(admin *Admins) *LoginMalicePrevent {
	ret := LoginMalicePrevent{}
	config.GetConfigField("admin", "login_malice_prevent").BindStruct(&ret)
	ret.admin = admin
	return &ret
}

// CheckLocked 返回账号解锁时间 为0时 即没有被锁或自动解锁
func (l *LoginMalicePrevent) CheckLocked() int {
	n, t := l.admin.CounterGet(CounterPassword)
	if n < l.Password || t == nil {
		return 0
	}
	return int(t.Add(time.Duration(l.LockTime) * time.Second).Sub(time.Now()).Seconds())
}
