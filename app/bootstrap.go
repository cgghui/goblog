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

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/ini.v1"
)

// Config App配置
type Config struct {
	ConfigFilePath string
	Log            *ini.Section
	Service        *ini.Section
}

// New 初始化app服务
func New(conf Config) {

	initDatabase()
	return

	cfg, err := ini.Load("../../config.ini", conf.ConfigFilePath)
	if err != nil {
		log.Printf("Fail load config file: %v\n", err)
		os.Exit(1)
	}
	conf.Log = cfg.Section("log")
	conf.Service = cfg.Section("Service")

	gin.DisableConsoleColor()

	router := gin.New()

	initLog(router, conf.Log)
	initRecovery(router, conf.Log)

	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "PONG")
	})

	s := &http.Server{
		Addr:         conf.Service.Key("listenAddr").MustString(""),
		Handler:      router,
		ReadTimeout:  time.Duration(conf.Service.Key("rtimeout").MustInt64(0)) * time.Second,
		WriteTimeout: time.Duration(conf.Service.Key("wtimeout").MustInt64(0)) * time.Second,
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

// User test
type User struct {
	ID        uint `gorm:"primary_key"`
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

// Profile test
type Profile struct {
	gorm.Model
	UserID   uint `gorm:"column:user_id"`
	Nickname string
	Age      string
	Sex      string
	User     User `gorm:"foreignkey:UserID"`
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

func initDatabase() {
	db, err := gorm.Open("mysql", "root:123123@(127.0.0.3)/stest?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Printf("Fail connect db: %v\n", err)
		return
	}
	// db.LogMode(true)
	// db.DB().SetMaxIdleConns(0)
	// db.DB().SetMaxOpenConns(0)
	defer db.Close()
	db.SingularTable(true)

	profile := User{}
	tmp := db.Model(&User{}).Related(&Profile{})
	tmp.First(&profile)
	fmt.Printf("%+v", profile)
}
