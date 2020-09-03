package system

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"strings"
	"time"
)

const (
	Name    = "GoBlog"               // 系统名称
	Version = "1.0.0"                // 系统版本
	HomeURL = "http://www.04559.com" // 系统官方网址
	Author  = "ChenGuangHui"         // 系统作者
)

var (
	DB    *gorm.DB
	Redis *redis.Client
)

// ConfMySQL MySQL配置参数
type ConfMySQL struct {
	Enable   bool            `ini:"enable"`
	Addr     string          `ini:"address"`
	User     string          `ini:"username"`
	Password string          `ini:"password"`
	Db       string          `ini:"dbname"`
	Prefix   string          `ini:"prefix"`
	Char     string          `ini:"charset"`
	MaxIdle  uint16          `ini:"max_idle"`
	MaxOpen  uint16          `ini:"max_open"`
	Log      string          `ini:"log"`
	LogLevel logger.LogLevel `ini:"log_level"`
}

// ConfRedis Redis配置参数
type ConfRedis struct {
	Enable  bool   `ini:"enable"`
	Address string `ini:"address"`
	Auth    string `ini:"auth"`
	DB      uint8  `ini:"use_db_index"`
}

// ConnectMySQL 连接MySQL 成功返回nil
func ConnectMySQL(c *ConfMySQL) error {
	if !c.Enable {
		return nil
	}
	if DB != nil {
	}
	if c.User == "" {
		return errors.New("MySQL conf incomplete")
	}
	if c.Db == "" {
		c.Db = c.User
	}
	if c.Addr == "" {
		c.Addr = "127.0.0.1:3306"
	} else {
		if strings.Index(c.Addr, ":") == -1 {
			c.Addr += ":3306"
		}
	}
	var output logger.Interface
	if c.Log == "" {
		output = logger.Default
	} else {
		f, err := os.OpenFile(c.Log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		output = logger.New(log.New(f, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold: time.Second,
			Colorful:      false,
			LogLevel:      c.LogLevel,
		})
	}
	db, err := gorm.Open(
		mysql.Open(fmt.Sprintf("%s:%s@(%s)/%s?charset=%s&parseTime=True&loc=Local", c.User, c.Password, c.Addr, c.Db, c.Char)),
		&gorm.Config{
			SkipDefaultTransaction: false,
			PrepareStmt:            true,
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   c.Prefix,
				SingularTable: true,
			},
			Logger: output,
		},
	)
	if err != nil {
		return fmt.Errorf("fail connect MySQL: %v", err)
	}
	x, err := db.DB()
	if err != nil {
		return fmt.Errorf("fail pool MySQL: %v", err)
	}
	x.SetMaxIdleConns(int(c.MaxIdle))
	x.SetMaxOpenConns(int(c.MaxOpen))
	DB = db
	return nil
}

// ConnRedis 连接Redis 成功返回nil
func ConnRedis(c *ConfRedis) error {
	if !c.Enable {
		return nil
	}
	if Redis != nil {
		_ = Redis.Close()
	}
	if c.Address == "" {
		c.Address = "127.0.0.1:6379"
	} else {
		if strings.Index(c.Address, ":") == -1 {
			c.Address += ":6379"
		}
	}
	Redis = redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     c.Address,
		Password: c.Auth,
		DB:       int(c.DB),
	})
	_, err := Redis.Ping(context.Background()).Result()
	return err
}
