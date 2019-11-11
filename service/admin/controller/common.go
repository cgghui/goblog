package controller

import (
	"goblog/app"

	"github.com/gin-gonic/gin"
)

// Common 全局通用
type Common struct {
}

//Construct 构造方法
func (c *Common) Construct(app *app.App) {

	app.Router.NoMethod(func(c *gin.Context) {
	})

	app.Router.NoRoute(func(c *gin.Context) {
	})

	app.Router.Static("/assets", "./assets")

}
