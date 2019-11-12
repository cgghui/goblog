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
func (c *Common) Construct(app *app.App) {

	app.Router.Static("/assets", "./view/static")

	app.Router.LoadHTMLFiles(
		"./view/index.html",
		"./view/error.html",
	)

	app.Router.NoMethod(func(c *gin.Context) {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"title": "Main website",
		})
	})

	app.Router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title": "Main website",
		})
	})

	app.Router.GET("/", func(ctx *gin.Context) {
		app.Output.DisplayHTML(ctx, "index.html")
	})

}
