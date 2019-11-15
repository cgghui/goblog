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

	appx.NoMethod(func(ctx *gin.Context) {
		app.Output.DisplayJSON(ctx, app.StatusForbidden, http.StatusForbidden)
	})

	appx.NoRoute(func(ctx *gin.Context) {
		app.Output.DisplayJSON(ctx, app.StatusNotFound, http.StatusNotFound)
	})

}
