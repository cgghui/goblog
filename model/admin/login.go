package admin

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"goblog/app"
	"goblog/model/config"
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"
)

// LoginCaptchaStatus 登录页是否须要验证码
// on  启用
// off 关闭
// condition 根据条件判断 这须要前端传入账号
func LoginCaptchaStatus() string {
	return config.GetConfigField("admin", "login_captcha").String()
}

// LoginCaptchaMemberStatus 登录账号是否须要验证码
// true 启用  false 关闭
func LoginCaptchaMemberStatus(a *Admins) bool {
	status := LoginCaptchaStatus()
	if status == "on" {
		return true
	}
	if status == "off" {
		return false
	}
	if a.CaptchaIsOpen == "Y" {
		return true
	}
	return NewLoginCaptchaCondition(NewLoginCounter(a.Username)).Check()
}

// < 登录密码校验 >
// 客户端将密码传输至服务端，是明文的，采用本机制可以进行加密传输，加密方式是RSA
// 实例NewLoginPasswordCrypt后，须要调用GenerateKey生成RSA密钥，并将返回的
// 的公钥传输给前端对密码进行加密。完成使用后，须要调用Clear()清除，否则密钥将一
// 直留存在Redis中。

// LoginPasswordCrypt 登录密码加密机制，采用RSA
type LoginPasswordCrypt struct {
	keyname  string
	password string
}

// NewLoginPasswordCrypt 实例
func NewLoginPasswordCrypt(a *Admins) *LoginPasswordCrypt {
	return &LoginPasswordCrypt{
		keyname:  "AdminTemp_" + a.Username,
		password: a.Password,
	}
}

// GenerateKey 创建RSA密钥 返回公钥
func (l *LoginPasswordCrypt) GenerateKey() string {
	PriKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	PubKey := &PriKey.PublicKey
	keyBytes, err := x509.MarshalPKIXPublicKey(PubKey)
	if err != nil {
		panic(err)
	}
	saved := app.RedisConn.HMSet(l.keyname, map[string]interface{}{
		"rsa_pubkey":    keyBytes,
		"rsa_prikey":    x509.MarshalPKCS1PrivateKey(PriKey),
		"rsa_create_at": time.Now().Unix(),
	})
	if saved.Err() != nil {
		panic(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: keyBytes}))
}

// Verify 密码检验
func (l *LoginPasswordCrypt) Verify(password string) (bool, error) {
	keyText, err := app.RedisConn.HGet(l.keyname, "rsa_prikey").Result()
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
	return PasswordVerify([]byte(l.password), pwd), nil
}

// Clear 清除
func (l *LoginPasswordCrypt) Clear() {
	app.RedisConn.HDel(l.keyname, "rsa_pubkey", "rsa_prikey", "rsa_create_at")
	return
}

// < 错误计数器 >
// 通过计数器，用户在登录时，可以对用户的错误行为进行计数，通过对错误次
// 数的控制，可有效的防止某此用户的恶意行为

// CounterField 错误行为的数据类型
type CounterField string

// 以下是对错误行为的定义，如果须要新增，在此处新增即可
var (
	LCP CounterField = "p" // 密码错误次数
	LCC CounterField = "c" // 验证码错误次数
)

// LoginCounter 登录错误记数器
type LoginCounter struct {
	keyname string
}

// NewLoginCounter 实例
func NewLoginCounter(username string) *LoginCounter {
	return &LoginCounter{
		keyname: "AdminsCounter_" + username,
	}
}

// Check 检查n是否大于field所记录的数，大于则返回true 否则返回false
func (lc *LoginCounter) Check(field CounterField, n int) bool {
	num, _ := lc.Get(field)
	return num > n
}

// Get 取出错误记录
func (lc *LoginCounter) Get(field CounterField) (int, *time.Time) {

	data, err := app.RedisConn.HGet(lc.keyname, string(field)).Result()
	if err != nil {
		return 0, nil
	}

	ret := make([]interface{}, 0)
	if err := json.Unmarshal([]byte(data), &ret); err != nil || len(ret) != 3 {
		lc.Clear(field)
		return 0, nil
	}

	if time.Unix(int64(ret[1].(float64)), 0).Sub(time.Now()).Seconds() <= 0 {
		lc.Clear(field)
		return 0, nil
	}

	t := time.Unix(int64(ret[2].(float64)), 0)

	return int(ret[0].(float64)), &t
}

// Incr 增加一次错误记录
func (lc *LoginCounter) Incr(field CounterField) {
	num, _ := lc.Get(field)
	num++
	data, err := json.Marshal([]interface{}{
		num,
		config.GetConfigField("admin", "login_counter_expire").TimeNowAddToUnix(),
		time.Now().Unix(),
	})
	if err != nil {
		panic(err)
	}
	if err := app.RedisConn.HSet(lc.keyname, string(field), data).Err(); err != nil {
		panic(err)
	}
	return
}

// Clear 清除错误记录
func (lc *LoginCounter) Clear(fields ...CounterField) {

	if len(fields) == 0 {
		if err := app.RedisConn.Del(lc.keyname).Err(); err != nil {
			panic(err)
		}
		return
	}

	fs := make([]string, 0)
	for i := 0; i < len(fields); i++ {
		if len(fields[i]) == 0 {
			continue
		}
		fs = append(fs, string(fields[i]))
	}
	if len(fs) == 0 {
		return
	}
	if err := app.RedisConn.HDel(lc.keyname, fs...).Err(); err != nil {
		panic(err)
	}
	return
}

// < 验证码启用条件 >
// 用户登录时，启用验证码的条件，写在此处的代码是用来决定用户是否须要输入验证码可
// 以在LoginCaptchaCondition这个结构体下增加方法，来做为新的条件，然后须要在
// Check方法中调用这个新方法，来做为最终的判定

// LoginCaptchaCondition 登录时启用验证码的条件 cinfig: login_captcha_condition
type LoginCaptchaCondition struct {
	Password int `json:"pwd_errn"`
	Captcha  int `json:"captcha_errn"`
	lc       *LoginCounter
}

// NewLoginCaptchaCondition 实例
func NewLoginCaptchaCondition(lc *LoginCounter) *LoginCaptchaCondition {
	ret := &LoginCaptchaCondition{}
	config.GetConfigField("admin", "login_captcha_condition").BindStruct(ret)
	ret.lc = lc
	return ret
}

// Check PasswordMax 和 CaptchaMax 只要有一个为true 返回true
func (l *LoginCaptchaCondition) Check() bool {
	return l.PasswordExceedMax() || l.CaptchaExceedMax()
}

// PasswordExceedMax 密码错误次数大于条件设定 返回true
func (l *LoginCaptchaCondition) PasswordExceedMax() bool {
	n, _ := l.lc.Get(LCP)
	return n > l.Password
}

// CaptchaExceedMax 验证码错误次数大于条件设定 返回true
func (l *LoginCaptchaCondition) CaptchaExceedMax() bool {
	n, _ := l.lc.Get(LCC)
	return n > l.Captcha
}

// Reset 将错误、验证码错误次数清零
// 注意：使用该方法，所有使用到LoginCounter的场景都会受到影响
func (l *LoginCaptchaCondition) Reset() {
	l.lc.Clear(LCP, LCC)
}

// < 安全防御 >
// 这是用于防止暴力破解密码的安全机制，当计数器中用户输入密码错误次数达到限制次数时，
// 用户账号将被在限制时间内锁定

// LoginMalicePrevent 管理员登录 密码错误次数限制（防止恶意尝试错误的密码）cinfig: login_malice_prevent
type LoginMalicePrevent struct {
	Password int   `json:"pwd_errn"`
	LockTime int64 `json:"lock_time"`
	lc       *LoginCounter
}

// NewLoginMalicePrevent 实例
func NewLoginMalicePrevent(lc *LoginCounter) *LoginMalicePrevent {
	ret := LoginMalicePrevent{}
	config.GetConfigField("admin", "login_malice_prevent").BindStruct(&ret)
	ret.lc = lc
	return &ret
}

// LockTTL 返回账号解锁时间 为0时 即没有被锁或自动解锁
func (l *LoginMalicePrevent) LockTTL() int {
	n, t := l.lc.Get(LCP)
	if n < l.Password || t == nil {
		return 0
	}
	return int(t.Add(time.Duration(l.LockTime) * time.Second).Sub(time.Now()).Seconds())
}

// Check 账号是否被锁定登录 true锁定 false没有锁定
func (l *LoginMalicePrevent) Check() bool {
	return l.LockTTL() != 0
}

// Unlock 解除锁定
func (l *LoginMalicePrevent) Unlock() {
	l.lc.Clear(LCP)
}

// <验证码>
// 验证码可有效防止机器人恶意攻击
// 1. 生成验证码
// 在取得LoginCaptcha的实例后，调用Generate生成验证码，如果验证码看不清楚，
// 将Generate返回的token转成keyid，然后再次调用Generate并将keyid传入，则
// 可生成新的验证码，token保持不变。

// LoginCaptcha 登录验证码
type LoginCaptcha struct {
	keyname string
	rsa     *RSA
}

// NewLoginCaptcha 实例
func NewLoginCaptcha(a *Admins) *LoginCaptcha {
	return &LoginCaptcha{
		keyname: "AdminTemp_" + a.Username,
		rsa:     a.RSA(),
	}
}

// Generate 加载验证码
// 如果传入keyids则在生成验证码时可保证token不变，注：此过程不验证keyid的合法性
// return1: 验证码图片 return2: 验证码token
func (l *LoginCaptcha) Generate(keyids ...string) (*string, string) {

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
		if err := app.RedisConn.HSet(l.keyname, hex.EncodeToString(keyid), verifyCode).Err(); err != nil {
			panic(err)
		}
		// 加密数据
		token, err := rsa.EncryptPKCS1v15(rand.Reader, l.rsa.PubKey, keyid)
		if err != nil {
			panic(err)
		}
		img := base64Captcha.CaptchaWriteToBase64Encoding(cimg)
		return &img, hex.EncodeToString(token)
	}
	// 刷新验证码
	if err := app.RedisConn.HSet(l.keyname, keyids[0], verifyCode).Err(); err != nil {
		panic(err)
	}
	// 加密TOKEN
	token, err := rsa.EncryptPKCS1v15(rand.Reader, l.rsa.PubKey, []byte(keyids[0]))
	if err != nil {
		panic(err)
	}
	img := base64Captcha.CaptchaWriteToBase64Encoding(cimg)
	return &img, hex.EncodeToString(token)
}

// Token2KeyID 将token转换为keyid
// 转换成功返回true 失败false
// 使用该方法转换成功后，传入的token将被转换，在该方法外部token会变成keyid
func (l *LoginCaptcha) Token2KeyID(token *string) bool {
	kid, err := hex.DecodeString(*token)
	if err != nil {
		return false
	}
	kid, err = rsa.DecryptPKCS1v15(rand.Reader, l.rsa.PriKey, kid)
	if err != nil {
		return false
	}
	*token = hex.EncodeToString(kid)
	return app.RedisConn.HExists(l.keyname, *token).Val()
}

// Verify 验证验证码
func (l *LoginCaptcha) Verify(code string, keyid string) (bool, error) {

	val, err := app.RedisConn.HGet(l.keyname, keyid).Result()
	if err != nil {
		return false, err
	}

	ret := make([]interface{}, 0)
	if err := json.Unmarshal([]byte(val), &ret); err != nil {
		l.Destroy(keyid)
		return false, err
	}

	if time.Unix(int64(ret[1].(float64)), 0).Sub(time.Now()).Seconds() <= 0 {
		l.Destroy(keyid)
		return false, nil
	}

	return strings.ToUpper(code) == strings.ToUpper(ret[0].(string)), nil
}

// Destroy 销毁验证码
func (l *LoginCaptcha) Destroy(keyid string) {
	app.RedisConn.HDel(l.keyname, keyid)
	return
}
