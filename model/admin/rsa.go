package admin

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"goblog/app"
)

// RSA 后台管理员RSA密钥 该密钥永久的保存 每个管理员拥有不同的证书
type RSA struct {
	Keyname string
	PriKey  *rsa.PrivateKey
	PubKey  *rsa.PublicKey
}

// NewRSA 实例管理员的RSA的操作对象
func NewRSA(admin *Admins) *RSA {
	return &RSA{
		Keyname: "AdminRsa_" + admin.Username,
	}
}

// CreateOrLoadKey 创建或加载RSA密钥
// 如果没密钥，则创建
// 如果有密钥，则加载
func (a *RSA) CreateOrLoadKey() {
	pemtext, err := app.RedisConn.HGetAll(a.Keyname).Result()
	if err != nil {
		panic(err)
	}
	if len(pemtext) == 0 {
		if err := a.CreateKey(); err != nil {
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

// CreateKey 创建RSA密钥 已存在则强制更换
func (a *RSA) CreateKey() error {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return err
	}
	a.PriKey = key
	a.PubKey = &key.PublicKey
	keyBytes, err := x509.MarshalPKIXPublicKey(a.PubKey)
	if err != nil {
		return err
	}
	saved := app.RedisConn.HMSet(a.Keyname, map[string]interface{}{
		"pubkey": keyBytes,
		"prikey": x509.MarshalPKCS1PrivateKey(a.PriKey),
	})
	if saved.Err() != nil {
		return err
	}
	return nil
}

// Encrypt 加密数据
func (a *RSA) Encrypt(data []byte) (string, error) {
	text, err := rsa.EncryptPKCS1v15(rand.Reader, a.PubKey, data)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(text), nil
}

// Decrypt 解密数据
func (a *RSA) Decrypt(hextext string) ([]byte, error) {
	data, err := hex.DecodeString(hextext)
	if err != nil {
		return nil, err
	}
	data, err = rsa.DecryptPKCS1v15(rand.Reader, a.PriKey, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
