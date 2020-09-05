package sys

import (
	"context"
	"fmt"
	"github.com/cgghui/goblog/sys/middleware"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
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
func New(rcs []RouteBuilder) *App {
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
	app := &App{Engine: gin.New()}
	if G.Listen == "" {
		log.Panicf("Fail listen address empty")
	}
	if G.Mode == "" {
		G.Mode = gin.DebugMode
	}
	gin.SetMode(G.Mode)
	app.Delims("{{", "}}")

	// middleware default
	app.Engine.Use(
		middlewareLogger(),
		middlewareRecovery(),
	)
	//middlewareCORS(app.Engine)
	//middlewareSession(app.Engine)
	return app
	//
	//for _, rc := range rcs {
	//	rc.Construct(app)
	//}
	//
	//s := &http.Server{
	//	Addr:         G.Listen,
	//	Handler:      app.Engine,
	//	ReadTimeout:  60 * time.Second,
	//	WriteTimeout: 60 * time.Second,
	//}
	//
	//if err := s.ListenAndServe(); err != nil {
	//	log.Printf("%v\n", err)
	//	os.Exit(1)
	//}
}

func (a *App) RegisterMiddleware(f ...gin.HandlerFunc) {
	a.Engine.Use(f...)
}

func (a *App) RegisterRouter() {

}

// 日志中间件
func middlewareLogger() gin.HandlerFunc {

	if G.WebLog == "" {
		return nil
	}

	output, err := os.OpenFile(G.WebLog, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Panicf("Fail open log file: [%s] %v", G.WebLog, err)
	}

	gin.DefaultWriter = io.MultiWriter(output)

	return gin.LoggerWithConfig(gin.LoggerConfig{
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
	})
}

// 异常恢复中件间
func middlewareRecovery() gin.HandlerFunc {

	if G.ErrLog == "" {
		return nil
	}

	output, err := os.OpenFile(G.ErrLog, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Panicf("Fail open log file: [%s] %v", G.ErrLog, err)
	}
	gin.DefaultErrorWriter = io.MultiWriter(output)

	return gin.Recovery()
}

//
//// 跨域中间件
//func middlewareCORS(router *gin.Engine) {
//	router.Use(cors.New(cors.Config{
//		AllowOrigins:     []string{"http://127.0.0.1:8787"},
//		AllowMethods:     []string{"GET", "POST", "OPTIONS", "DELETE", "PUT"},
//		AllowHeaders:     []string{"Origin", "AccessToken"},
//		ExposeHeaders:    []string{"Content-Length"},
//		AllowCredentials: true,
//		MaxAge:           24 * time.Hour,
//	}))
//}

// session中间件
func middleSession(router *gin.Engine) gin.HandlerFunc {
	if !G.S.Enable {
		return nil
	}
	// 解析Redis链接
	target, err := url.Parse(G.S.SaveHandler)
	if err != nil {
		log.Panicf("session save handler parse: %v", err)
	}
	if target.Scheme != "redis" {
		log.Panic("session save handler not redis")
	}
	opt := redis.Options{
		Network: "tcp",
	}
	if target.Hostname() == "" {
		opt.Addr = "127.0.0.1:6379"
	} else {
		if target.Port() == "" {
			opt.Addr = target.Hostname() + ":6379"
		} else {
			opt.Addr = target.Host
		}
	}
	arg := target.Query()
	if auth := arg.Get("auth"); auth != "" {
		opt.Password = auth
	}
	if db := arg.Get("db"); db != "" {
		idx, err := strconv.Atoi(db)
		if err != nil {
			log.Panic("session redis select db not int")
		}
		opt.DB = idx
	}
	conn := redis.NewClient(&opt)
	_, err = Redis.Ping(context.Background()).Result()
	if err != nil {
		log.Panicf("Fail SESSION store redis: %v", err)
	}
	if G.S.Name == "" {
		G.S.Name = "GOBLOG_SESSION"
	}
	store := middleware.NewStorageToRedis(conn, []byte(G.S.Secret))
	// 数据序列化方式
	switch G.S.DataSerialize {
	case "gob":
		store.SetSerializer(middleware.GobSerializer{})
	case "json":
		store.SetSerializer(middleware.JSONSerializer{})
	default:
		store.SetSerializer(middleware.GobSerializer{})
	}
	//
	store.SetKeyPrefix(G.S.Prefix)
	// Cookie的有效域名
	if G.S.CookieDomain != "" {
		store.Opts.Domain = G.S.CookieDomain
	}
	// Cookie的有效路径
	if G.S.CookiePath == "" {
		G.S.CookiePath = "/"
	}
	store.Opts.Path = G.S.CookiePath
	// Cookie的有效期
	if G.S.CookieExpires <= 0 {
		G.S.CookieExpires = 2592000
	}
	store.Opts.MaxAge = G.S.CookieExpires
	//
	return sessions.Sessions(G.S.Name, store)
}
