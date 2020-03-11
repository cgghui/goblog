package admin

import (
	"goblog/app"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type Gender string

func (g Gender) In() bool {
	for _, gender := range Genders {
		if gender == g {
			return true
		}
	}
	return false
}

var Genders = []Gender{"M", "W", "X"}

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
	Username      string `gorm:"type:varchar(64)"`
	Password      string `gorm:"type:varchar(256)"`
	Status        string `gorm:"type:enum('locked','normal')"`
	CaptchaStatus string `gorm:"type:enum('Y','C')"`
	LoginIP       string `gorm:"type:varchar(32)"`
	Nickname      string `gorm:"type:varchar(64)"`
	Gender        string `gorm:"type:enum('M','W', 'X')"`
	Avatar        string `gorm:"type:varchar(128)"`
	Mobile        string `gorm:"type:varchar(32)"`
	Email         string `gorm:"type:varchar(64)"`
	Intro         string `gorm:"type:varchar(256)"`
	rsa           *RSA
}

// GetByUsername 根据管理员姓名，取出该管理员对象
func GetByUsername(username string) *Admins {
	admin := &Admins{}
	app.DBConn.Where("username = ?", username).First(admin)
	return admin
}

// NicknameCheckUsed 如果未使用返回0，否则返回占用的ID
func NicknameCheckUsed(nickname string) uint {
	admin := &Admins{}
	app.DBConn.Where("nickname = ?", nickname).First(admin)
	return admin.ID
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

// Output 可以展示到前端的数据
func (a *Admins) Output() map[string]interface{} {
	return map[string]interface{}{
		"uid":             a.ID,
		"username":        a.Username,
		"nickname":        a.Nickname,
		"gender":          a.Gender,
		"avatar":          a.Avatar,
		"mobile":          a.Mobile,
		"email":           a.Email,
		"intro":           a.Intro,
		"join_time":       a.CreatedAt.Unix(),
		"last_login_time": a.UpdatedAt.Unix(),
	}
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
