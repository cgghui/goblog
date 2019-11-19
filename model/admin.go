package model

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// Admins 全局参数配置
type Admins struct {
	gorm.Model
	Nickname       string `gorm:"type:varchar(128)"`
	Username       string `gorm:"type:varchar(128)"`
	Password       string `gorm:"type:varchar(256)"`
	Status         string `gorm:"type:enum('locked','normal')"`
	CaptchaIsOpen  string `gorm:"type:enum('Y','C')"`
	GoogleAuthCode string `gorm:"type:varchar(256)"`
	LoginIP        string `gorm:"type:varchar(32)"`
}

// AdminsLog 全局参数配置
type AdminsLog struct {
	ID            uint      `gorm:"primary_key"`
	LoginUID      uint      `gorm:"type:int(10);column:login_uid"`
	LoginUsername string    `gorm:"type:varchar(128)"`
	VisitDatetime time.Time `gorm:"type:datetime"`
	IP            string    `gorm:"type:varchar(64)"`
	Action        string    `gorm:"type:varchar(32)"`
	Msg           string    `gorm:"type:varchar(256)"`
	Info          string    `gorm:"type:varchar(1024)"`
	CreatedAt     time.Time `gorm:"type:timestamp"`
}

// AdminsErrorCounter 登录时操作错误记数器
type AdminsErrorCounter struct {
	*Admins
	Password       uint
	Captcha        uint
	GoogleAuthCode uint
}

// AdminLoginCaptchaCheck 检查后台管理员登录是否须要使用验证码
// return true开启 false关闭
func AdminLoginCaptchaCheck(admin ...*Admins) bool {

	v := GetConfigField("admin", "login_captcha").String()

	if v == "on" {
		return true
	}

	if v == "off" || len(admin) == 0 {
		return false
	}

	// 按条件判定是否须要验证码
	if admin[0].CaptchaIsOpen == "Y" {
		return true
	}

	// ......

	return false
}

// AdminGeneratePassword 生成新的后台管理员登录密码
func AdminGeneratePassword(pwd string) string {
	temp, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("Error: model.AdminGeneratePassword %v", err))
	}
	return string(temp)
}

// AdminVerifyPassword 密码检验
func AdminVerifyPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
