package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"gopkg.in/ini.v1"

	// MySQL Driver
	_ "github.com/go-sql-driver/mysql"
)

// App App配置
type App struct {
	Config map[string]*ini.Section
	Router *gin.Engine
	DB     *gorm.DB
}

// New 初始化app服务
func New() {

	cfg, err := ini.Load("../../config.ini", "config.ini")
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
		},
		Router: gin.New(),
	}

	initLog(app.Router, app.Config["log"])
	initRecovery(app.Router, app.Config["log"])
	initDatabase(app)

	app.Router.GET("/ping", func(c *gin.Context) {

		c.String(200, "PONG")
	})

	s := &http.Server{
		Addr:         app.Config["service"].Key("listenAddr").MustString(""),
		Handler:      app.Router,
		ReadTimeout:  time.Duration(app.Config["service"].Key("rtimeout").MustInt64(0)) * time.Second,
		WriteTimeout: time.Duration(app.Config["service"].Key("wtimeout").MustInt64(0)) * time.Second,
	}

	s.ListenAndServe()
}

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

func initDatabase(app *App) {

	conf := app.Config["db"]

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
	defer db.Close()
	//db.LogMode(true)
	db.DB().SetMaxIdleConns(conf.Key("max_idle").MustInt(9))
	db.DB().SetMaxOpenConns(conf.Key("max_open").MustInt(0))
	db.SingularTable(true)

	app.DB = db
}
