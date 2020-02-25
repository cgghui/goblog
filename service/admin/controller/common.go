package controller

import (
	"goblog/app"
	"goblog/model/admin"
	"goblog/model/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SessionUser 当前登录用户的信息
var SessionUser admin.LoginSessionData

// Common 全局通用
type Common struct {
}

//Construct 构造方法
func (c *Common) Construct(appx *app.App) {

	appx.NoMethod(func(ctx *gin.Context) {
		app.Output().DisplayJSON(ctx, app.StatusForbidden, http.StatusForbidden)
	})

	appx.NoRoute(func(ctx *gin.Context) {
		app.Output().DisplayJSON(ctx, app.StatusNotFound, http.StatusNotFound)
	})

	appx.Use(checkSession([]string{
		"/auth/check",
		"/auth/load_captcha",
		"/auth/passport",
	}))

}

func checkSession(skipRoutes []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, r := range skipRoutes {
			if r == ctx.Request.URL.Path {
				return
			}
		}
		tk := config.GetConfigField("admin", "session_name").String()
		tk = ctx.Request.Header.Clone().Get(tk)
		auth := admin.NewLogin(ctx.ClientIP())
		if !auth.Token2KeyID(&tk) {
			app.Output(gin.H{"tip": "SessionID无效"}).DisplayJSON(ctx, app.StatusAuthInvalid)
			ctx.Abort()
			return
		}
		if !auth.Check(tk, &SessionUser) {
			app.Output(gin.H{"tip": "您掉线了"}).DisplayJSON(ctx, app.StatusAuthInvalid)
			ctx.Abort()
			return
		}
	}
}
