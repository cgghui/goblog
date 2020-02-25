package controller

import (
	"fmt"
	"goblog/app"
	"goblog/model/config"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var tplvars = gin.H{
	"sysn":      app.SystemName,
	"sysv":      app.SystemVersion,
	"sysAuthor": app.SystemAuthor,
	"sysURL":    app.SystemHomeURL,
}

// Common 全局通用
type Common struct {
}

// TPLRootPath 模板目录根路径
const TPLRootPath string = "./view/tpl/"

//Construct 构造方法
func (c *Common) Construct(appx *app.App) {

	// 静态服务
	SSU := app.SysConf["service"].Key("staticServiceURL").MustString("")
	SSP := app.SysConf["service"].Key("staticServicePath").MustString("")
	if SSU != "" && SSP != "" {
		appx.Static(SSU, SSP)
	}

	// 全局模板变量
	tplvars["surl"] = SSU
	tplvars["container_name"] = app.SysConf["service"].Key("frontend_ContainerName").MustString("")
	tplvars["front_end_version"] = app.SysConf["service"].Key("frontend_Version").MustString("")
	tplvars["session_name"] = config.GetConfigField("admin", "session_name").String()

	// 加载模板文件
	appx.LoadHTMLGlob(TPLRootPath + "*")

	//
	appx.NoMethod(func(ctx *gin.Context) {
		app.Output(tplvars).DisplayHTML(ctx, "error_403.html", http.StatusForbidden)
	})

	//
	appx.NoRoute(func(ctx *gin.Context) {
		app.Output(tplvars).DisplayHTML(ctx, "error_404.html", http.StatusNotFound)
	})

	// 加载首页
	appx.GET("/", func(ctx *gin.Context) {
		app.Output(tplvars).DisplayHTML(ctx, "index.html")
	})

	// 加载各个模板文件
	appx.GET("/tpl/:dir/:file", func(ctx *gin.Context) {

		ctx.Header("Content-Type", "text/plain; charset=utf-8")

		dir, tplf := ctx.Param("dir"), ctx.Param("file")
		if len(dir) == 0 || len(tplf) == 0 {
			app.Output(tplvars).DisplayHTML(ctx, "error_404.html", http.StatusNotFound)
			return
		}

		tplf = dir + "_" + tplf

		if _, err := os.Stat(TPLRootPath + tplf); err != nil {
			ctx.Error(fmt.Errorf(`tpl /view/tpl/%s %v`, tplf, err))
			app.Output(tplvars).DisplayHTML(ctx, "error_403.html", http.StatusForbidden)
			return
		}

		app.Output(tplvars).DisplayHTML(ctx, tplf)
	})

	// 加载配置文件 js
	appx.GET("/config.js", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "application/javascript")
		app.Output(tplvars).DisplayHTML(ctx, "config.tpl.js")
	})
}
