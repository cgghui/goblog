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

// SessionID 就是KeyID
var SessionID string

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

	skipAuths := make([]string, 0)
	config.Get("admin", "skip_auths").BindStruct(&skipAuths)

	appx.Use(c.checkSession(skipAuths))

}

func (*Common) checkSession(skipRoutes []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, r := range skipRoutes {
			if r == ctx.Request.URL.Path {
				return
			}
		}
		tk := config.Get("admin", "session_name").String()
		tk = ctx.Request.Header.Clone().Get(tk)
		if len(tk) < 8 {
			app.Output(gin.H{"tip": "SessionID无效 Err1"}).DisplayJSON(ctx, app.StatusAuthInvalid)
			ctx.Abort()
			return
		}
		auth := admin.NewLogin(ctx.ClientIP())
		if !auth.Token2KeyID(&tk) {
			app.Output(gin.H{"tip": "SessionID无效 Err2"}).DisplayJSON(ctx, app.StatusAuthInvalid)
			ctx.Abort()
			return
		}
		if !auth.Check(tk, &SessionUser) {
			app.Output(gin.H{"tip": "您掉线了"}).DisplayJSON(ctx, app.StatusAuthInvalid)
			ctx.Abort()
			return
		}
		SessionID = tk
	}
}
