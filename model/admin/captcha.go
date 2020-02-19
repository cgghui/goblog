package admin

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"goblog/app"
	"goblog/model/config"
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"
)

// CaptchaLoginCheck 如果登录须要验证码返回true
func (a *Admins) CaptchaLoginCheck() bool {
	return AdminLoginCaptchaCheck(a)
}

// CaptchaLaod 加载验证码 return1: 验证码图片 return2: 验证码令牌
func (a *Admins) CaptchaLaod(keyids ...string) (*string, string) {

	conf := base64Captcha.ConfigCharacter{}
	config.GetConfigField("admin", "login_captcha_config").BindStruct(&conf)

	cimg := base64Captcha.EngineCharCreate(conf)
	verifyCode, err := json.Marshal([]interface{}{
		cimg.VerifyValue,
		config.GetConfigField("admin", "login_captcha_expire").TimeNowAddToUnix(),
	})
	if err != nil {
		panic(err)
	}
	// 生成验证码
	if len(keyids) == 0 {
		md5ctx := md5.New()
		md5ctx.Write(verifyCode)
		keyid := md5ctx.Sum(nil)
		if err := app.RedisConn.HSet(a.tempkey(), hex.EncodeToString(keyid), verifyCode).Err(); err != nil {
			panic(err)
		}
		// 加密数据
		token, err := rsa.EncryptPKCS1v15(rand.Reader, a.PubKey, keyid)
		if err != nil {
			panic(err)
		}
		img := base64Captcha.CaptchaWriteToBase64Encoding(cimg)
		return &img, hex.EncodeToString(token)
	}
	// 刷新验证码
	if err := app.RedisConn.HSet(a.tempkey(), keyids[0], verifyCode).Err(); err != nil {
		panic(err)
	}
	img := base64Captcha.CaptchaWriteToBase64Encoding(cimg)
	return &img, ""
}

// CaptchaTokenCheck token检测
func (a *Admins) CaptchaTokenCheck(token *string) bool {
	tk, err := hex.DecodeString(*token)
	if err != nil {
		return false
	}
	tk, err = rsa.DecryptPKCS1v15(rand.Reader, a.PriKey, tk)
	if err != nil {
		return false
	}
	*token = hex.EncodeToString(tk)
	return app.RedisConn.HExists(a.tempkey(), *token).Val()
}

// CaptchaVerify 验证验证码
func (a *Admins) CaptchaVerify(code string, keyid string) (bool, error) {

	token, err := hex.DecodeString(keyid)
	if err != nil {
		return false, err
	}

	token, err = rsa.DecryptPKCS1v15(rand.Reader, a.PriKey, token)
	if err != nil {
		return false, err
	}

	val, err := app.RedisConn.HGet(a.tempkey(), hex.EncodeToString(token)).Result()
	if err != nil {
		return false, err
	}

	ret := make([]interface{}, 0)
	if err := json.Unmarshal([]byte(val), &ret); err != nil {
		a.CaptchaDestroy(keyid)
		return false, err
	}

	if time.Unix(int64(ret[1].(float64)), 0).Sub(time.Now()).Seconds() <= 0 {
		a.CaptchaDestroy(keyid)
		return false, nil
	}

	return strings.ToUpper(code) == strings.ToUpper(ret[0].(string)), nil
}

// CaptchaDestroy 销毁验证码
func (a *Admins) CaptchaDestroy(keyid string) {
	app.RedisConn.HDel(a.tempkey(), keyid)
	return
}
