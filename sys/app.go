package sys

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

// RouteBuilder 路由构造器
type RouteBuilder interface {
	Construct(*App)
}

// App App配置
type App struct {
	*gin.Engine
}

// New 初始化app服务
func New(rcs []RouteBuilder) {

	if DB != nil {
		db, _ := DB.DB()
		if err := db.Ping(); err != nil {
			log.Panicf("db ping error %v", err)
		}
		defer db.Close()
	}

	if Redis != nil {
		if _, err := Redis.Ping(context.Background()).Result(); err != nil {
			log.Panicf("Fail connect redis: %v", err)
		}
		defer Redis.Close()
	}

	gin.DisableConsoleColor()

	app := &App{gin.New()}

	if G.Listen == "" {
		log.Panicf("Fail listen address empty")
	}

	if G.Mode == "" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(G.Mode)
	}
	app.Delims("{[", "]}")

	// middleware
	middlewareLog(app.Engine)
	middlewareRecovery(app.Engine)
	middlewareCORS(app.Engine)
	middlewareSession(app.Engine)

	//
	for _, rc := range rcs {
		rc.Construct(app)
	}

	s := &http.Server{
		Addr:         G.Listen,
		Handler:      app.Engine,
		ReadTimeout:  time.Duration(SysConf["service"].Key("rtimeout").MustInt64(0)) * time.Second,
		WriteTimeout: time.Duration(SysConf["service"].Key("wtimeout").MustInt64(0)) * time.Second,
	}

	if err := s.ListenAndServe(); err != nil {
		log.Printf("%v\n", err)
		os.Exit(1)
	}
}

// 日志中间件
func middlewareLog(router *gin.Engine) {

	logf := SysConf["log"].Key("wwwlog").MustString("")
	if logf == "" {
		return
	}

	logfp, err := os.OpenFile(logf, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Printf("Fail open log file: [%s] %v\n", logf, err)
		os.Exit(1)
	}

	gin.DefaultWriter = io.MultiWriter(logfp)

	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("%s - [%s] \"%s %s %s\" %d \"%s\" \"%s\" \"%s\"\n",
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

// 异常恢复中件间
func middlewareRecovery(router *gin.Engine) {

	logf := SysConf["log"].Key("recovery").MustString("")
	if logf == "" {
		return
	}

	logfp, err := os.OpenFile(logf, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Printf("Fail open log file: [%s] %v\n", logf, err)
		os.Exit(1)
	}
	gin.DefaultErrorWriter = io.MultiWriter(logfp)
	router.Use(gin.Recovery())

	return
}

// 跨域中间件
func middlewareCORS(router *gin.Engine) {
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://127.0.0.1:8787"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "DELETE", "PUT"},
		AllowHeaders:     []string{"Origin", "AccessToken"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}))
}

// session中间件
func middlewareSession(router *gin.Engine) {

	conf := SysConf["session"]
	if conf.Key("status").MustString("enable") == "disable" {
		return
	}

	store, err := redis.NewStore(
		conf.Key("max_idle").MustInt(10),
		conf.Key("network").MustString("tcp"),
		conf.Key("host").MustString("127.0.0.1:6379"),
		conf.Key("auth").MustString(""),
		[]byte(conf.Key("secret").MustString("")),
	)
	if err != nil {
		log.Printf("Fail init session: %v\n", err)
		os.Exit(1)
	}

	Session = store

	router.Use(sessions.Sessions(conf.Key("name").MustString("SESSION"), store))
}
