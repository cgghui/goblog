package adminsys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"goblog/app"
	"time"

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

// PasswordVerify treturn true 正对, false 错误
func PasswordVerify(hashedPassword, password []byte) bool {
	return bcrypt.CompareHashAndPassword(hashedPassword, password) == nil
}

// PasswordVerify 密码检验
// 该方法佼验证密码 须要使用 CreatePasswordEncryptRSA 的公钥对密码加密
func (a *Admins) PasswordVerify(password string) (bool, error) {
	keyText, err := app.RedisConn.HGet(a.tempkey(), "rsa_prikey").Result()
	if err != nil {
		return false, err
	}
	PriKey, err := x509.ParsePKCS1PrivateKey([]byte(keyText))
	if err != nil {
		return false, err
	}
	temp, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return false, err
	}
	pwd, err := rsa.DecryptPKCS1v15(rand.Reader, PriKey, temp)
	if err != nil {
		return false, err
	}
	return PasswordVerify([]byte(a.Password), pwd), nil
}

// PasswordCreateEncryptRSA 创建RSA密钥 返回公钥
func (a *Admins) PasswordCreateEncryptRSA() string {
	PriKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	PubKey := &PriKey.PublicKey
	keyBytes, err := x509.MarshalPKIXPublicKey(PubKey)
	if err != nil {
		panic(err)
	}
	saved := app.RedisConn.HMSet(a.tempkey(), map[string]interface{}{
		"rsa_pubkey":    keyBytes,
		"rsa_prikey":    x509.MarshalPKCS1PrivateKey(PriKey),
		"rsa_create_at": time.Now().Unix(),
	})
	if saved.Err() != nil {
		panic(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: keyBytes}))
}
