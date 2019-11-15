package controller

import (
	"fmt"
	"goblog/app"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Common 全局通用
type Common struct {
}

// TPLRootPath 模板目录根路径
const TPLRootPath string = "./view/tpl/"

//Construct 构造方法
func (c *Common) Construct(appx *app.App) {

	SSU := app.SysConf["service"].Key("staticServiceURL").MustString("")
	SSP := app.SysConf["service"].Key("staticServicePath").MustString("")
	if SSU != "" && SSP != "" {
		appx.Static(SSU, SSP)
	}

	appx.LoadHTMLGlob(TPLRootPath + "*")

	app.Output.Assgin("sysn", app.SystemName)
	app.Output.Assgin("sysv", app.SystemVersion)
	app.Output.Assgin("sysAuthor", app.SystemAuthor)
	app.Output.Assgin("sysURL", app.SystemHomeURL)
	app.Output.Assgin("surl", SSU)
	app.Output.Assgin("container_name", app.SysConf["service"].Key("frontend_ContainerName").MustString(""))
	app.Output.Assgin("front_end_version", app.SysConf["service"].Key("frontend_Version").MustString(""))

	appx.NoMethod(func(ctx *gin.Context) {
		app.Output.DisplayHTML(ctx, "error_403.html", http.StatusForbidden)
	})

	appx.NoRoute(func(ctx *gin.Context) {
		app.Output.DisplayHTML(ctx, "error_404.html", http.StatusNotFound)
	})

	appx.GET("/", func(ctx *gin.Context) {
		app.Output.DisplayHTML(ctx, "index.html")
	})

	appx.GET("/tpl/:dir/:file", func(ctx *gin.Context) {

		ctx.Header("Content-Type", "text/plain; charset=utf-8")

		dir, tplf := ctx.Param("dir"), ctx.Param("file")
		if len(dir) == 0 || len(tplf) == 0 {
			app.Output.DisplayHTML(ctx, "error_404.html", http.StatusNotFound)
			return
		}

		tplf = dir + "_" + tplf

		if _, err := os.Stat(TPLRootPath + tplf); err != nil {
			ctx.Error(fmt.Errorf(`tpl /view/tpl/%s %v`, tplf, err))
			app.Output.DisplayHTML(ctx, "error_403.html", http.StatusForbidden)
			return
		}

		app.Output.DisplayHTML(ctx, tplf)
	})

	appx.GET("/config.js", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "application/javascript")
		app.Output.DisplayHTML(ctx, "config.tpl.js")
	})
}
