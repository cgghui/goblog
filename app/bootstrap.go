package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/ini.v1"
)

// Bootstrap 结构
type Bootstrap struct {
	Router *gin.Engine
	Config string
}

// New 初始化app服务
func New(conf Config) {

	cfg, err := ini.Load(conf.ConFile)
	if err != nil {
		log.Printf("Fail load config file: %v", err)
		os.Exit(1)
	}

	bt := &Bootstrap{
		Router: gin.New(),
	}

	if logf := cfg.Section("log").Key("wwwlog").MustString(""); logf != "" {
		initLog(bt.Router, logf)
	}

	if logf := cfg.Section("log").Key("recovery").MustString(""); logf != "" {
		logfp, err := os.OpenFile(logf, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Printf("Fail open log file: [%s] %v", logf, err)
			os.Exit(1)
		}
		gin.DefaultErrorWriter = io.MultiWriter(logfp)
		bt.Router.Use(gin.Recovery())
	}

	bt.Router.GET("/ping", func(c *gin.Context) {
		if 1 == 1 {
			panic("data Error")
		}
		c.String(200, "PONG")
	})
	s := &http.Server{
		Addr:         conf.ListenAddr,
		Handler:      bt.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		// MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}

func initLog(router *gin.Engine, logf string) {

	gin.DisableConsoleColor()

	logfp, err := os.OpenFile(logf, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Printf("Fail open log file: [%s] %v", logf, err)
		os.Exit(1)
	}

	gin.DefaultWriter = io.MultiWriter(logfp)

	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf(`%s - [%s] "%s %s %s" %d "%s" "%s" "%s"\n`,
				param.ClientIP,
				param.TimeStamp.Format("2006-01-02 15:04:05"),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Latency,
				param.Request.UserAgent(),
				param.ErrorMessage,
			)
		},
	}))

	return
}
