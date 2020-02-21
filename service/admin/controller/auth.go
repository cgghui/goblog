package controller

import (
	"goblog/app"
	"goblog/model/admin"
	"time"

	"github.com/gin-gonic/gin"
)

// Auth 授权
type Auth struct {
}

//Construct 构造方法
func (a *Auth) Construct(app *app.App) {
	app.GET("/auth/check", check)
	app.GET("/auth/load_captcha", loadCaptcha)
	app.POST("/auth/passport", passport)
}

// 登录前检测账号的状态
// 最好是在账号输入框失去焦点时 调用接口进行检测
// URL auth/check?username=val
// Method GET
func check(ctx *gin.Context) {

	// 检测输入的账号
	username := ctx.Query("username")
	if len(username) == 0 {
		app.Output(gin.H{"tip": "请输入账号"}).DisplayJSON(ctx, app.StatusQueryInvalid)
		return
	}

	// 从数据库取账号信息
	adminuser := admin.GetByUsername(username)
	if !adminuser.Has() {
		app.Output(gin.H{"username": username}).DisplayJSON(ctx, app.StatusUserNotExist)
		return
	}

	// 输出至浏览器的数据
	data := gin.H{
		"locked":          false,
		"unlock_ttl":      0,
		"pubkey":          "",
		"captcha_is_open": false,
		"captcha":         gin.H{},
	}

	lc := admin.NewLoginCounter(username)

	// 如果账号被锁定
	if s := admin.NewLoginMalicePrevent(lc).LockTTL(); s != 0 {
		data["locked"] = true
		data["unlock_ttl"] = s
		app.Output(data).DisplayJSON(ctx, app.StatusOK)
		return
	}

	// 取加密密码明文的RSA公钥
	data["pubkey"] = admin.NewLoginPasswordCrypt(adminuser).GenerateKey()

	// 如果须要验证码
	if admin.NewLoginCaptchaCondition(lc).Check() {
		captcha := admin.NewLoginCaptcha(adminuser)
		img, token := captcha.Generate()
		data["captcha_is_open"] = true
		data["captcha"] = gin.H{"image": *img, "token": token}
	}

	app.Output(data).DisplayJSON(ctx, app.StatusOK)
}

// 登录时重新加载验证码 如果用户看清 可调用该接口刷新验证码
// URL auth/loadCaptcha?username=val&token=
// Method GET
func loadCaptcha(ctx *gin.Context) {

	username := ctx.Query("username")
	if len(username) == 0 {
		app.Output(gin.H{"tip": "请传入账号"}).DisplayJSON(ctx, app.StatusQueryInvalid)
		return
	}

	adminuser := admin.GetByUsername(username)

	if !adminuser.Has() {
		app.Output(gin.H{"tip": "无效账号"}).DisplayJSON(ctx, app.StatusQueryInvalid)
		return
	}

	token := ctx.Query("token")
	if len(token) == 0 {
		app.Output(gin.H{"tip": "请传入令牌"}).DisplayJSON(ctx, app.StatusQueryInvalid)
		return
	}

	captcha := admin.NewLoginCaptcha(adminuser)

	if !captcha.Token2KeyID(&token) {
		app.Output(gin.H{"tip": "传入令牌的无效"}).DisplayJSON(ctx, app.StatusQueryInvalid)
		return
	}

	img, _ := captcha.Generate(token)
	app.Output(gin.H{"image": *img}).DisplayJSON(ctx, app.StatusOK)
}

func passport(ctx *gin.Context) {

	var form admin.FormLogin
	var captcha *admin.LoginCaptcha
	var captchaOpend bool

	if err := ctx.ShouldBind(&form); err != nil {
		ctx.Error(err)
		app.Output(gin.H{"tip": "参数无效"}).DisplayJSON(ctx, app.StatusQueryInvalid)
		return
	}

	adminuser := form.GetAdmin()

	if !adminuser.Has() {
		app.Output(gin.H{"username": form.Username}).DisplayJSON(ctx, app.StatusUserNotExist)
		return
	}

	nowtime := time.Now()

	logtext := &admin.AdminsLog{
		LoginUID:      adminuser.ID,
		LoginUsername: adminuser.Username,
		IP:            ctx.ClientIP(),
		Action:        "LOGIN",
		Msg:           "登录失败",
		VisitDatetime: nowtime,
	}
	defer func() {
		go app.DBConn.Create(logtext)
	}()

	lc := admin.NewLoginCounter(adminuser.Username)

	if s := admin.NewLoginMalicePrevent(lc).LockTTL(); s != 0 {
		logtext.Msg = "登录失败: 账号被锁定"
		app.Output(gin.H{"locked": true, "unlock_ttl": s}).DisplayJSON(ctx, app.StatusUserLocked)
		return
	}

	if admin.NewLoginCaptchaCondition(lc).Check() {
		captcha = admin.NewLoginCaptcha(adminuser)
		captchaOpend = true
	}

	if captchaOpend {
		if !form.CheckCaptchaQuantity() {
			lc.Incr(admin.LCC)
			logtext.Msg = "登录失败: 验证码无效(码数不一致)"
			app.Output(gin.H{"captcha_code": form.CaptchaC, "ret": -1}).DisplayJSON(ctx, app.StatusCaptchaError)
			return
		}
		keyid := form.CaptchaT
		if !captcha.Token2KeyID(&keyid) {
			lc.Incr(admin.LCC)
			logtext.Msg = "登录失败: 验证码无效(令牌是无效)"
			app.Output(gin.H{"captcha_code": form.CaptchaC, "ret": -2}).DisplayJSON(ctx, app.StatusCaptchaError)
			return
		}
		ok, err := captcha.Verify(form.CaptchaC, keyid)
		if err != nil {
			ctx.Error(err)
		}
		if !ok {
			lc.Incr(admin.LCC)
			logtext.Msg = "登录失败: 验证码错误"
			app.Output(gin.H{"captcha_code": form.CaptchaC, "ret": -3}).DisplayJSON(ctx, app.StatusCaptchaError)
			return
		}
	}

	password := admin.NewLoginPasswordCrypt(adminuser)

	ok, err := password.Verify(form.Password)
	if err != nil {
		ctx.Error(err)
	}
	if !ok {
		lc.Incr(admin.LCP)
		output := app.Output()
		if admin.NewLoginCaptchaCondition(lc).Check() {
			captcha = admin.NewLoginCaptcha(adminuser)
			captchaOpend = true
		}
		output.Assgin("captcha_open", captchaOpend)
		if captchaOpend {
			var (
				img   *string
				token string
			)
			keyid := form.CaptchaT
			if captcha.Token2KeyID(&keyid) {
				img, token = captcha.Generate(keyid)
			} else {
				img, token = captcha.Generate()
			}
			output.Assgin("captcha_token", token)
			output.Assgin("captcha_image", *img)
		}
		logtext.Msg = "登录失败: 密码错误"
		output.DisplayJSON(ctx, app.StatusPasswordErr)
		return
	}

	lc.Clear()
	password.Clear()

	logtext.Msg = "登录成功"
	app.Output(gin.H{"access_token": ""}).DisplayJSON(ctx, app.StatusOK)
}
