package model

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
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
	PriKey         *rsa.PrivateKey
	PubKey         *rsa.PublicKey
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

// 管理员登录时的错误记录字段
var (
	AdminCounterPassword       = "password"
	AdminCounterCaptcha        = "captcha"
	AdminCounterGoogleAuthCode = "google_authcode"
)

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

	ret := ConfigAdminLCC{}
	GetConfigField("admin", "login_captcha_condition").BindJSON(&ret)

	// 如果有符合验证码开启条件的 则开启
	n1 := admin[0].ErrorCounterGet(AdminCounterPassword)
	n2 := admin[0].ErrorCounterGet(AdminCounterCaptcha)
	n3 := admin[0].ErrorCounterGet(AdminCounterGoogleAuthCode)
	if n1 > ret.Password || n2 > ret.Captcha || n3 > ret.GoogleAuthCode {
		return true
	}

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

// Has 管理员是否存在
func (a *Admins) Has() bool {
	return a.ID != 0
}

// BuildKeyToRSA 创建RSA密钥 永久性使用
func (a *Admins) BuildKeyToRSA() {
	k := "AdminsCert_" + a.Username
	pemtext, err := app.RedisConn.HGetAll(k).Result()
	if err != nil {
		panic(fmt.Sprintf("Error: model.admins.BuildKeyToRSA Redis %v", err))
	}
	if len(pemtext) == 0 {
		key, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			panic(fmt.Sprintf("Error: model.admins.BuildKeyToRSA Generate RSA key %v", err))
		}
		a.PriKey = key
		a.PubKey = &key.PublicKey

		keyBytes, err := x509.MarshalPKIXPublicKey(a.PubKey)
		if err != nil {
			panic(fmt.Sprintf("Error: model.admins.BuildKeyToRSA build RSA public key %v", err))
		}

		saved := app.RedisConn.HMSet(k, map[string]interface{}{
			"pubkey": keyBytes,
			"prikey": x509.MarshalPKCS1PrivateKey(a.PriKey),
		})
		if saved.Err() != nil {
			panic(fmt.Sprintf("Error: model.admins.BuildKeyToRSA RSA key saved fail %v", err))
		}
	} else {
		pubkey, err := x509.ParsePKIXPublicKey([]byte(pemtext["pubkey"]))
		if err != nil {
			panic(fmt.Sprintf("Error: model.admins.BuildKeyToRSA parse public key fail %v", err))
		}
		prikey, err := x509.ParsePKCS1PrivateKey([]byte(pemtext["prikey"]))
		if err != nil {
			panic(fmt.Sprintf("Error: model.admins.BuildKeyToRSA parse private key fail %v", err))
		}
		a.PubKey = pubkey.(*rsa.PublicKey)
		a.PriKey = prikey
	}
}

// GetPasswordEncryptRSA 创建RSA密钥 临时使用
func (a *Admins) GetPasswordEncryptRSA() string {

	k := "AdminsCertTemp_" + a.Username

	PriKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(fmt.Sprintf("Error: model.admins.GetPasswordEncryptRSA Generate RSA key %v", err))
	}
	PubKey := &PriKey.PublicKey

	keyBytes, err := x509.MarshalPKIXPublicKey(PubKey)
	if err != nil {
		panic(fmt.Sprintf("Error: model.admins.GetPasswordEncryptRSA build RSA public key %v", err))
	}

	saved := app.RedisConn.HMSet(k, map[string]interface{}{
		"pubkey":    keyBytes,
		"prikey":    x509.MarshalPKCS1PrivateKey(PriKey),
		"create_at": time.Now().Unix(),
	})
	if saved.Err() != nil {
		panic(fmt.Sprintf("Error: model.admins.GetPasswordEncryptRSA RSA key saved fail %v", err))
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: keyBytes}))
}

// VerifyPassword 密码检验
func (a *Admins) VerifyPassword(password string) bool {
	return AdminVerifyPassword(a.Password, password)
}

// ErrorCounterGet 取出错误记录
func (a *Admins) ErrorCounterGet(field string) int {

	data, err := app.RedisConn.HGet(a.eckey(), field).Result()
	if err != nil {
		return 0
	}

	ret := make([]interface{}, 0)
	if err := json.Unmarshal([]byte(data), &ret); err != nil || len(ret) != 2 {
		return 0
	}

	s := time.Unix(int64(ret[1].(float64)), 0).Sub(time.Now()).Seconds()
	if s <= 0 {
		a.ErrorCounterClear(field)
		return 0
	}

	return int(ret[0].(float64))
}

// ErrorCounterIncr 增加一次错误记录
func (a *Admins) ErrorCounterIncr(field string) {
	ttl := time.Duration(GetConfigField("admin", "login_counter_expire").Int()) * time.Second
	data, err := json.Marshal([]interface{}{a.ErrorCounterGet(field) + 1, time.Now().Add(ttl).Unix()})
	if err != nil {
		panic(fmt.Sprintf("Error: model.admins.ErrorCounterIncr %v", err))
	}
	if err := app.RedisConn.HSet(a.eckey(), field, data).Err(); err != nil {
		panic(fmt.Sprintf("Error: model.admins.ErrorCounterIncr %v", err))
	}
	return
}

// ErrorCounterClear 清除错误记录
func (a *Admins) ErrorCounterClear(fields ...string) {
	// 删除所有
	if len(fields) == 0 {
		if err := app.RedisConn.Del(a.eckey()).Err(); err != nil {
			panic(fmt.Sprintf("Error: model.admins.ErrorCounterClear %v", err))
		}
		return
	}
	// 删除指定
	if err := app.RedisConn.HDel(a.eckey(), fields...).Err(); err != nil {
		panic(fmt.Sprintf("Error: model.admins.ErrorCounterClear %v", err))
	}
	return
}

func (a *Admins) eckey() string {
	return "AdminsErrorCounter_" + a.Username
}
