package app

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	rediscli "github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
	"gopkg.in/ini.v1"

	// MySQL Driver
	_ "github.com/go-sql-driver/mysql"
)

// 系统预设参数
const (
	SystemName    = "GoBlog"               // 系统名称
	SystemVersion = "1.0.0"                // 系统版本
	SystemHomeURL = "http://www.04559.com" // 系统官方网址
	SystemAuthor  = "chen guang hui"       // 系统作者
)

// Global Var
var (
	SysConf   map[string]*ini.Section
	DBConn    *gorm.DB
	RedisConn *rediscli.Client
	Session   redis.Store
	Output    *OutputFMT
)

func init() {

	cfg, err := ini.Load("../../config.ini", "config.ini")
	if err != nil {
		log.Printf("Fail load config file: %v\n", err)
		os.Exit(1)
	}

	SysConf = map[string]*ini.Section{
		"log":     cfg.Section("log"),
		"service": cfg.Section("service"),
		"db":      cfg.Section("MySQL"),
		"redis":   cfg.Section("Redis"),
		"session": cfg.Section("Session"),
	}

	initDatabase()
	if DBConn != nil {
		if err := DBConn.DB().Ping(); err != nil {
			log.Printf("Fail database error: %v\n", err)
			os.Exit(1)
		}
		defer DBConn.Close()
	}

	initRedis()
	if RedisConn != nil {
		if _, err := RedisConn.Ping().Result(); err != nil {
			log.Printf("Fail connect redis: %v\n", err)
			os.Exit(1)
		}
		defer RedisConn.Close()
	}

	Output = &OutputFMT{H: gin.H{}}
}

func initDatabase() {

	conf := SysConf["db"]
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

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return conf.Key("prefix").MustString("gb_") + defaultTableName
	}

	DBConn = db
}

func initRedis() {
	conf := SysConf["redis"]
	if conf.Key("status").MustString("enable") == "disable" {
		return
	}
	RedisConn = rediscli.NewClient(&rediscli.Options{
		Network:  conf.Key("network").MustString("tcp"),
		Addr:     conf.Key("host").MustString("127.0.0.1:6379"),
		Password: conf.Key("auth").MustString(""),
		DB:       conf.Key("index").MustInt(0),
	})
}
