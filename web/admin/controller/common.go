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

	appx.Router.Static("/assets", "./view/static")

	appx.Router.LoadHTMLFiles(
		"./view/index.html",
		"./view/error_403.html",
		"./view/error_404.html",
	)

	appx.Output.Assgin("sysn", app.SystemName)
	appx.Output.Assgin("sysv", app.SystemVersion)

	appx.Router.NoMethod(func(ctx *gin.Context) {
		appx.Output.DisplayHTML(ctx, "error_403.html", http.StatusForbidden)
	})

	appx.Router.NoRoute(func(ctx *gin.Context) {
		appx.Output.DisplayHTML(ctx, "error_404.html", http.StatusNotFound)
	})

	appx.Router.GET("/", func(ctx *gin.Context) {
		appx.Output.DisplayHTML(ctx, "index.html")
	})

}
