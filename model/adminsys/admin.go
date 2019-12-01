package adminsys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"goblog/app"
	"goblog/model/config"
	"time"

	"github.com/jinzhu/gorm"
)

// Admins 全局参数配置
type Admins struct {
	gorm.Model
	Nickname      string `gorm:"type:varchar(128)"`
	Username      string `gorm:"type:varchar(128)"`
	Password      string `gorm:"type:varchar(256)"`
	Status        string `gorm:"type:enum('locked','normal')"`
	CaptchaIsOpen string `gorm:"type:enum('Y','C')"`
	LoginIP       string `gorm:"type:varchar(32)"`
	PriKey        *rsa.PrivateKey
	PubKey        *rsa.PublicKey
}

// AdminsLog 全局参数配置
type AdminsLog struct {
	ID            uint      `gorm:"primary_key"`
	LoginUID      uint      `gorm:"type:int(10);column:login_uid"`
	LoginUsername string    `gorm:"type:varchar(128)"`
	IP            string    `gorm:"type:varchar(64)"`
	Action        string    `gorm:"type:varchar(32)"`
	Msg           string    `gorm:"type:varchar(256)"`
	Info          string    `gorm:"type:varchar(1024)"`
	VisitDatetime time.Time `gorm:"type:datetime"`
}

// AdminLoginCaptchaCheck return true open, false close
// 检查登录时是否须要使用验证码
func AdminLoginCaptchaCheck(admin ...*Admins) bool {

	v := config.GetConfigField("admin", "login_captcha").String()

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

	return NewLoginCaptchaCondition(admin[0]).Check()
}

// Has 管理员是否存在 如果返回true则存在 否则不存在
func (a *Admins) Has() bool {
	return a.ID != 0
}

// CheckLocked 是否账户是否被锁定 如果锁定返回大于0的数 未锁定返回0
func (a *Admins) CheckLocked() int {
	return NewLoginMalicePrevent(a).CheckLocked()
}

// BuildKeyToRSA 创建RSA密钥 永久性使用
func (a *Admins) BuildKeyToRSA() {
	k := "AdminsRsa_" + a.Username
	pemtext, err := app.RedisConn.HGetAll(k).Result()
	if err != nil {
		panic(err)
	}
	if len(pemtext) == 0 {
		key, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			panic(err)
		}
		a.PriKey = key
		a.PubKey = &key.PublicKey

		keyBytes, err := x509.MarshalPKIXPublicKey(a.PubKey)
		if err != nil {
			panic(err)
		}

		saved := app.RedisConn.HMSet(k, map[string]interface{}{
			"pubkey": keyBytes,
			"prikey": x509.MarshalPKCS1PrivateKey(a.PriKey),
		})
		if saved.Err() != nil {
			panic(err)
		}
	} else {
		pubkey, err := x509.ParsePKIXPublicKey([]byte(pemtext["pubkey"]))
		if err != nil {
			panic(err)
		}
		prikey, err := x509.ParsePKCS1PrivateKey([]byte(pemtext["prikey"]))
		if err != nil {
			panic(err)
		}
		a.PubKey = pubkey.(*rsa.PublicKey)
		a.PriKey = prikey
	}
}

// ClearTemp 清除临时数据
func (a *Admins) ClearTemp() {
	app.RedisConn.Del(a.tempkey())
	return
}

func (a *Admins) tempkey() string {
	return "AdminTemp_" + a.Username
}
