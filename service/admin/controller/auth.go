package controller

import (
	"goblog/app"
	"goblog/model/adminsys"

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
	admin := &adminsys.Admins{}
	app.DBConn.Where("username = ?", username).First(admin)

	// 如果账号不存在
	if !admin.Has() {
		app.Output(gin.H{"username": username}).DisplayJSON(ctx, app.StatusUserNotExist)
		return
	}

	admin.BuildKeyToRSA()

	// 输出至浏览器的数据
	data := gin.H{
		"locked":          false,
		"unlock_ttl":      0,
		"pubkey":          "",
		"captcha_is_open": false,
		"captcha":         gin.H{},
	}

	// 如果账号被锁定
	if s := admin.CheckLocked(); s != 0 {
		data["locked"] = true
		data["unlock_ttl"] = s
		app.Output(data).DisplayJSON(ctx, app.StatusOK)
		return
	}

	// 取加密密码明文的RSA公钥
	data["pubkey"] = admin.PasswordCreateEncryptRSA()

	// 如果须要验证码
	if admin.CaptchaLoginCheck() {
		img, token := admin.CaptchaLaod()
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

	admin := &adminsys.Admins{}
	app.DBConn.Where("username = ?", username).First(admin)

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

	var form adminsys.FormAdminLogin

	if ctx.ShouldBind(&form) != nil {
		app.Output(gin.H{"tip": "参数无效"}).DisplayJSON(ctx, app.StatusQueryInvalid)
		return
	}

	admin := form.GetAdmin()

	if !admin.Has() {
		app.Output(gin.H{"username": form.Username}).DisplayJSON(ctx, app.StatusUserNotExist)
		return
	}

	admin.BuildKeyToRSA()

	if s := admin.CheckLocked(); s != 0 {
		app.Output(gin.H{"locked": true, "unlock_ttl": s}).DisplayJSON(ctx, app.StatusUserLocked)
		return
	}

	isOpenCaptcha := admin.CaptchaLoginCheck()

	if isOpenCaptcha {
		if !form.CheckCaptchaQuantity() {
			admin.CounterIncr(adminsys.CounterCaptcha)
			app.Output(gin.H{"captcha_code": form.CaptchaCode, "ret": -1}).DisplayJSON(ctx, app.StatusCaptchaError)
			return
		}
		ok, err := admin.CaptchaVerify(form.CaptchaCode, form.CaptchaToken)
		if err != nil {
			ctx.Error(err)
		}
		if !ok {
			admin.CounterIncr(adminsys.CounterCaptcha)
			app.Output(gin.H{"captcha_code": form.CaptchaCode, "ret": -2}).DisplayJSON(ctx, app.StatusCaptchaError)
			return
		}
	}

	ok, err := admin.PasswordVerify(form.Password)
	if err != nil {
		ctx.Error(err)
	}
	if !ok {
		admin.CounterIncr(adminsys.CounterPassword)
		output := app.Output()
		output.Assgin("captcha_open", isOpenCaptcha)
		if isOpenCaptcha {
			token := form.CaptchaToken
			var img *string
			if admin.CaptchaTokenCheck(&token) {
				img, _ = admin.CaptchaLaod(form.CaptchaToken)
				token = form.CaptchaToken
			} else {
				img, token = admin.CaptchaLaod()
			}
			output.Assgin("captcha_token", token)
			output.Assgin("captcha_image", *img)
		}
		output.DisplayJSON(ctx, app.StatusPasswordErr)
		return
	}

	admin.CounterClear()
	admin.ClearTemp()
	app.Output(gin.H{"access_token": ""}).DisplayJSON(ctx, app.StatusOK)
}
