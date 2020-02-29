package controller

import (
	"goblog/app"

	"github.com/gin-gonic/gin"
)

// Notice 通知
type Notice struct {
}

//Construct 构造方法
func (n *Notice) Construct(app *app.App) {
	auth := app.Group("/notice")
	auth.GET("/new_message_check", n.newMessageCheck)
}

func (n *Notice) newMessageCheck(ctx *gin.Context) {
	app.Output(gin.H{
		"newmsg": true,
	}).DisplayJSON(ctx, app.StatusOK)
}
