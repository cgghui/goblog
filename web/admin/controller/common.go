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

	SSU := appx.Config["service"].Key("staticServiceURL").MustString("")
	SSP := appx.Config["service"].Key("staticServicePath").MustString("")
	if SSU != "" && SSP != "" {
		appx.Router.Static(SSU, SSP)
	}

	appx.Router.LoadHTMLGlob(TPLRootPath + "*")

	appx.Output.Assgin("sysn", app.SystemName)
	appx.Output.Assgin("sysv", app.SystemVersion)
	appx.Output.Assgin("surl", SSU)
	appx.Output.Assgin("container_name", appx.Config["service"].Key("frontend_ContainerName").MustString(""))
	appx.Output.Assgin("front_end_version", appx.Config["service"].Key("frontend_Version").MustString(""))

	appx.Router.NoMethod(func(ctx *gin.Context) {
		appx.Output.DisplayHTML(ctx, "error_403.html", http.StatusForbidden)
	})

	appx.Router.NoRoute(func(ctx *gin.Context) {
		appx.Output.DisplayHTML(ctx, "error_404.html", http.StatusNotFound)
	})

	appx.Router.GET("/", func(ctx *gin.Context) {
		appx.Output.DisplayHTML(ctx, "index.html")
	})

	appx.Router.GET("/tpl/:dir/:file", func(ctx *gin.Context) {

		ctx.Header("Content-Type", "text/plain; charset=utf-8")

		dir, tplf := ctx.Param("dir"), ctx.Param("file")
		if len(dir) == 0 || len(tplf) == 0 {
			appx.Output.DisplayHTML(ctx, "error_404.html", http.StatusNotFound)
			return
		}

		tplf = dir + "_" + tplf

		if _, err := os.Stat(TPLRootPath + tplf); err != nil {
			ctx.Error(fmt.Errorf(`tpl /view/tpl/%s %v`, tplf, err))
			appx.Output.DisplayHTML(ctx, "error_403.html", http.StatusForbidden)
			return
		}

		appx.Output.DisplayHTML(ctx, tplf)
	})

	appx.Router.GET("/config.js", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "application/javascript")
		appx.Output.DisplayHTML(ctx, "config.tpl.js")
	})
}
