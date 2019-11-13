package controller

import (
	"goblog/app"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Common 全局通用
type Common struct {
}

//Construct 构造方法
func (c *Common) Construct(appx *app.App) {

	SSU := appx.Config["service"].Key("staticServiceURL").MustString("")
	SSP := appx.Config["service"].Key("staticServicePath").MustString("")
	if SSU != "" && SSP != "" {
		appx.Router.Static(SSU, SSP)
	}

	appx.Router.LoadHTMLFiles(
		"./view/index.html",
		"./view/error_403.html",
		"./view/error_404.html",
		"./view/config.tpl.js",
	)

	appx.Output.Assgin("sysn", app.SystemName)
	appx.Output.Assgin("sysv", app.SystemVersion)
	appx.Output.Assgin("surl", SSU)
	appx.Output.Assgin("box_name", appx.Config["service"].Key("containerBoxName").MustString(""))

	appx.Router.NoMethod(func(ctx *gin.Context) {
		appx.Output.DisplayHTML(ctx, "error_403.html", http.StatusForbidden)
	})

	appx.Router.NoRoute(func(ctx *gin.Context) {
		appx.Output.DisplayHTML(ctx, "error_404.html", http.StatusNotFound)
	})

	appx.Router.GET("/", func(ctx *gin.Context) {
		appx.Output.DisplayHTML(ctx, "index.html")
	})

	appx.Router.GET("/config.js", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "application/javascript")
		appx.Output.DisplayHTML(ctx, "config.tpl.js")
	})
}
