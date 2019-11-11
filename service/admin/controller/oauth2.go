package controller

import (
	"errors"
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
	app.Router.POST("/oauth2/authorize", o.authorize)
	app.Router.GET("/oauth2/token", o.token)
}

// AuthorizeInput 授权提交的内容
type AuthorizeInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (o *Oauth2) authorize(c *gin.Context) {
	input := AuthorizeInput{}
	if err := c.BindJSON(&input); err != nil {
		c.Error(errors.New("无效参数"))
		c.AsciiJSON(http.StatusOK, gin.H{"status": false, "code": 100, "msg": "无效参数"})
		return
	}
}

func (o *Oauth2) token(c *gin.Context) {

}
