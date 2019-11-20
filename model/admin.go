package model

import (
	"encoding/json"
	"fmt"
	"goblog/app"
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

// VerifyPassword 密码检验
func (a *Admins) VerifyPassword(password string) bool {
	return AdminVerifyPassword(a.Password, password)
}

// ErrorCounterGet 取出错误记录
func (a *Admins) ErrorCounterGet(field string) int {
	// 从Redis获取数据失败
	data := app.RedisConn.HGet(a.eckey(), field).String()

	// 将数据解析为JSON失败
	ret := make([]interface{}, 0)
	if err := json.Unmarshal([]byte(data), &ret); err != nil || len(ret) != 2 {
		return 0
	}
	// 第1个索引位置的值不是一个时间戳
	et, ok := ret[1].(int64)
	if !ok {
		return 0
	}
	// 限制已经失效，重0开始计数
	if time.Unix(et, 0).Sub(time.Now()).Seconds() <= 0 {
		return 0
	}
	n, ok := ret[0].(int)
	if !ok {
		return 0
	}
	return n
}

// ErrorCounterIncr 增加一次错误记录
func (a *Admins) ErrorCounterIncr(field string) {
	data, err := json.Marshal([]interface{}{a.ErrorCounterGet(field) + 1, time.Now().Unix()})
	if err != nil {
		panic(fmt.Sprintf("Error: model.ErrorCounterIncr %v", err))
	}
	if err := app.RedisConn.HSet(a.eckey(), field, data).Err(); err != nil {
		panic(fmt.Sprintf("Error: model.ErrorCounterIncr %v", err))
	}
	return
}

// ErrorCounterDecr 减去一次错误记录
func (a *Admins) ErrorCounterDecr(field string) {
	n := a.ErrorCounterGet(field) - 1
	if n <= 0 {
		n = 0
	}
	data, err := json.Marshal([]interface{}{n, time.Now().Unix()})
	if err != nil {
		panic(fmt.Sprintf("Error: model.ErrorCounterDecr %v", err))
	}
	if err := app.RedisConn.HSet(a.eckey(), field, data).Err(); err != nil {
		panic(fmt.Sprintf("Error: model.ErrorCounterDecr %v", err))
	}
	return
}

// ErrorCounterClear 清除错误记录
func (a *Admins) ErrorCounterClear(fields ...string) {
	// 删除所有
	if len(fields) == 0 {
		if err := app.RedisConn.Del(a.eckey()).Err(); err != nil {
			panic(fmt.Sprintf("Error: model.ErrorCounterClear %v", err))
		}
		return
	}
	// 删除指定
	if err := app.RedisConn.HDel(a.eckey(), fields...).Err(); err != nil {
		panic(fmt.Sprintf("Error: model.ErrorCounterClear %v", err))
	}
	return
}

func (a *Admins) eckey() string {
	return "AdminsErrorCounter_" + a.Username
}
