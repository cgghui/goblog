package admin

import (
	"goblog/app"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// PasswordGenerate retun 密码密文
func PasswordGenerate(password string) string {
	temp, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(temp)
}

// PasswordVerify return true 正对, false 错误
func PasswordVerify(hashedPassword, password []byte) bool {
	return bcrypt.CompareHashAndPassword(hashedPassword, password) == nil
}

// Admins 全局参数配置
type Admins struct {
	gorm.Model
	Nickname      string `gorm:"type:varchar(128)"`
	Username      string `gorm:"type:varchar(128)"`
	Password      string `gorm:"type:varchar(256)"`
	Status        string `gorm:"type:enum('locked','normal')"`
	CaptchaIsOpen string `gorm:"type:enum('Y','C')"`
	LoginIP       string `gorm:"type:varchar(32)"`
	rsa           *RSA
}

// GetByUsername 根据管理员姓名，取出该管理员对象
func GetByUsername(username string) *Admins {
	admin := &Admins{}
	app.DBConn.Where("username = ?", username).First(admin)
	return admin
}

// GetByID 根据管理员ID，取出该管理员对象
func GetByID(id uint) *Admins {
	admin := &Admins{}
	app.DBConn.Where("id = ?", id).First(admin)
	return admin
}

// Has 管理员是否存在 true存在 false不存在
func (a *Admins) Has() bool {
	return a.ID != 0
}

// RSA 加密解密操作对象
// 每个账号持有的密钥不同，因此账号与账号间不可混淆使用
func (a *Admins) RSA() *RSA {
	if a.rsa != nil {
		return a.rsa
	}
	a.rsa = NewRSA(a)
	a.rsa.CreateOrLoadKey()
	return a.rsa
}
