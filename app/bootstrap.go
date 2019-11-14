package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	rediscli "github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
	"gopkg.in/ini.v1"

	// MySQL Driver
	_ "github.com/go-sql-driver/mysql"
)

const (

	// SystemName 系统名称
	SystemName = "GoBlog"

	// SystemVersion 系统版本
	SystemVersion = "1.0.0"

	// SystemHomeURL 系统URL地址
	SystemHomeURL = "http://www.04559.com"

	// SystemAuthor 系统作者
	SystemAuthor = "chen guang hui"
)

// RouteBuilder 路由构造器
type RouteBuilder interface {
	Construct(*App)
}

// App App配置
type App struct {
	Config  map[string]*ini.Section
	Router  *gin.Engine
	DB      *gorm.DB
	Redis   *rediscli.Client
	Session redis.Store
	Output  *Output
}

// New 初始化app服务
func New(cfp string, rcs []RouteBuilder) {

	cfg, err := ini.Load(cfp, "config.ini")
	if err != nil {
		log.Printf("Fail load config file: %v\n", err)
		os.Exit(1)
	}

	gin.DisableConsoleColor()

	app := &App{
		Config: map[string]*ini.Section{
			"log":     cfg.Section("log"),
			"service": cfg.Section("service"),
			"db":      cfg.Section("MySQL"),
			"redis":   cfg.Section("Redis"),
			"session": cfg.Section("Session"),
		},
		Router: gin.New(),
	}

	listenAddr := app.Config["service"].Key("listenAddr").MustString("")
	if listenAddr == "" {
		log.Printf("Fail listen address empty\n")
		os.Exit(1)
	}

	gin.SetMode(app.Config["service"].Key("mode").MustString(gin.DebugMode))

	app.Router.Delims("{[", "]}")

	// log and recovery
	initLog(app.Router, app.Config["log"])
	initRecovery(app.Router, app.Config["log"])

	// database MySQL
	initDatabase(app)
	if app.DB != nil {
		if err := app.DB.DB().Ping(); err != nil {
			log.Printf("Fail database error: %v\n", err)
			os.Exit(1)
		}
		defer app.DB.Close()
	}

	// cache Redis
	initRedis(app)
	if app.Redis != nil {
		if _, err := app.Redis.Ping().Result(); err != nil {
			log.Printf("Fail connect redis: %v\n", err)
			os.Exit(1)
		}
		defer app.Redis.Close()
	}

	// Session
	initSession(app)

	// 输出至浏览器
	initOutput(app)

	//
	for _, rc := range rcs {
		rc.Construct(app)
	}

	s := &http.Server{
		Addr:         listenAddr,
		Handler:      app.Router,
		ReadTimeout:  time.Duration(app.Config["service"].Key("rtimeout").MustInt64(0)) * time.Second,
		WriteTimeout: time.Duration(app.Config["service"].Key("wtimeout").MustInt64(0)) * time.Second,
	}

	s.ListenAndServe()
}

// initLog 初始化日志
func initLog(router *gin.Engine, conf *ini.Section) {

	logf := conf.Key("wwwlog").MustString("")
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

// initRecovery 初始化异常恢复
func initRecovery(router *gin.Engine, conf *ini.Section) {

	logf := conf.Key("recovery").MustString("")
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

// initDatabase 初始化数据库ORM
func initDatabase(app *App) {

	conf := app.Config["db"]
	if conf.Key("status").MustString("enable") == "disable" {
		return
	}

	dbuser := conf.Key("user").MustString("")
	if dbuser == "" {
		log.Printf("Fail db user empty\n")
		os.Exit(1)
	}

	db, err := gorm.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@(%s)/%s?charset=%s&parseTime=True&loc=Local",
			dbuser,
			conf.Key("password").MustString(""),
			conf.Key("host").MustString("127.0.0.1:3306"),
			conf.Key("dbname").MustString(dbuser),
			conf.Key("charset").MustString("utf8"),
		),
	)
	if err != nil {
		log.Printf("Fail connect db: %v\n", err)
		os.Exit(1)
	}

	db.DB().SetMaxIdleConns(conf.Key("max_idle").MustInt(9))
	db.DB().SetMaxOpenConns(conf.Key("max_open").MustInt(0))
	db.SingularTable(true)

	app.DB = db
}

func initRedis(app *App) {
	conf := app.Config["redis"]
	if conf.Key("status").MustString("enable") == "disable" {
		return
	}
	app.Redis = rediscli.NewClient(&rediscli.Options{
		Network:  conf.Key("network").MustString("tcp"),
		Addr:     conf.Key("host").MustString("127.0.0.1:6379"),
		Password: conf.Key("auth").MustString(""),
		DB:       conf.Key("index").MustInt(0),
	})
}

func initSession(app *App) {
	conf := app.Config["session"]
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
	app.Session = store
	app.Router.Use(sessions.Sessions(conf.Key("name").MustString("SESSION"), store))
}
