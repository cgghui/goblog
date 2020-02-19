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
	adminuser := admin.GetAdminByUsernme(username)
	if !adminuser.Has() {
		app.Output(gin.H{"username": username}).DisplayJSON(ctx, app.StatusUserNotExist)
		return
	}

	adminuser.BuildKeyToRSA()

	// 输出至浏览器的数据
	data := gin.H{
		"locked":          false,
		"unlock_ttl":      0,
		"pubkey":          "",
		"captcha_is_open": false,
		"captcha":         gin.H{},
	}

	// 如果账号被锁定
	if s := adminuser.CheckLocked(); s != 0 {
		data["locked"] = true
		data["unlock_ttl"] = s
		app.Output(data).DisplayJSON(ctx, app.StatusOK)
		return
	}

	// 取加密密码明文的RSA公钥
	data["pubkey"] = adminuser.PasswordCreateEncryptRSA()

	// 如果须要验证码
	if adminuser.CaptchaLoginCheck() {
		img, token := adminuser.CaptchaLaod()
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

	admin := admin.GetAdminByUsernme(username)

	if !admin.Has() {
		app.Output(gin.H{"tip": "无效账号"}).DisplayJSON(ctx, app.StatusQueryInvalid)
		return
	}

	admin.BuildKeyToRSA()

	token := ctx.Query("token")
	if len(token) == 0 {
		app.Output(gin.H{"tip": "请传入令牌"}).DisplayJSON(ctx, app.StatusQueryInvalid)
		return
	}

	if !admin.CaptchaTokenCheck(&token) {
		app.Output(gin.H{"tip": "传入令牌的无效"}).DisplayJSON(ctx, app.StatusQueryInvalid)
		return
	}

	img, _ := admin.CaptchaLaod(token)
	app.Output(gin.H{"image": *img}).DisplayJSON(ctx, app.StatusOK)
}

func passport(ctx *gin.Context) {

	var form admin.FormAdminLogin

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

	adminuser.BuildKeyToRSA()

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

	if s := adminuser.CheckLocked(); s != 0 {
		logtext.Msg = "登录失败: 账号被锁定"
		app.Output(gin.H{"locked": true, "unlock_ttl": s}).DisplayJSON(ctx, app.StatusUserLocked)
		return
	}

	isOpenCaptcha := adminuser.CaptchaLoginCheck()

	if isOpenCaptcha {
		if !form.CheckCaptchaQuantity() {
			adminuser.CounterIncr(admin.CounterCaptcha)
			logtext.Msg = "登录失败: 验证码错误(-1)"
			app.Output(gin.H{"captcha_code": form.CaptchaCode, "ret": -1}).DisplayJSON(ctx, app.StatusCaptchaError)
			return
		}
		ok, err := adminuser.CaptchaVerify(form.CaptchaCode, form.CaptchaToken)
		if err != nil {
			ctx.Error(err)
		}
		if !ok {
			adminuser.CounterIncr(admin.CounterCaptcha)
			logtext.Msg = "登录失败: 验证码错误(-2)"
			app.Output(gin.H{"captcha_code": form.CaptchaCode, "ret": -2}).DisplayJSON(ctx, app.StatusCaptchaError)
			return
		}
	}

	ok, err := adminuser.PasswordVerify(form.Password)
	if err != nil {
		ctx.Error(err)
	}
	if !ok {
		adminuser.CounterIncr(admin.CounterPassword)
		output := app.Output()
		output.Assgin("captcha_open", isOpenCaptcha)
		if isOpenCaptcha {
			token := form.CaptchaToken
			var img *string
			if adminuser.CaptchaTokenCheck(&token) {
				img, _ = adminuser.CaptchaLaod(token)
				token = form.CaptchaToken
			} else {
				img, token = adminuser.CaptchaLaod()
			}
			output.Assgin("captcha_token", token)
			output.Assgin("captcha_image", *img)
		}
		logtext.Msg = "登录失败: 密码错误"
		output.DisplayJSON(ctx, app.StatusPasswordErr)
		return
	}

	adminuser.CounterClear()
	adminuser.ClearTemp()

	logtext.Msg = "登录成功"
	app.Output(gin.H{"access_token": ""}).DisplayJSON(ctx, app.StatusOK)
}
