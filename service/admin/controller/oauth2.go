package controller

import (
	"goblog/app"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Oauth2 授权
type Oauth2 struct {
	*app.App
}

//Construct 构造方法
func (o *Oauth2) Construct(app *app.App) {
	app.Router.GET("/oauth2/authorize", o.authorize)
	app.Router.GET("/oauth2/token", o.token)
}

func (o *Oauth2) authorize(c *gin.Context) {
	c.String(http.StatusOK, "wait.....")
}

func (o *Oauth2) token(c *gin.Context) {

}
