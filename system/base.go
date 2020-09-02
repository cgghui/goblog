package system

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
	"strings"

	// MySQL Driver
	_ "github.com/go-sql-driver/mysql"
)

const (
	Name    = "GoBlog"               // 系统名称
	Version = "1.0.0"                // 系统版本
	HomeURL = "http://www.04559.com" // 系统官方网址
	Author  = "ChenGuangHui"         // 系统作者
)

var (
	DB       *gorm.DB
	dbDebug  *gorm.DB
	dbNormal *gorm.DB
	Redis    *redis.Client
)

// ConfMySQL MySQL配置参数
type ConfMySQL struct {
	Enable   bool   `ini:"enable"`
	Address  string `ini:"address"`
	Username string `ini:"username"`
	Password string `ini:"password"`
	Dbname   string `ini:"dbname"`
	Prefix   string `ini:"prefix"`
	Charset  string `ini:"charset"`
	MaxIdle  uint16 `ini:"max_idle"`
	MaxOpen  uint16 `ini:"max_open"`
	Debug    bool   `ini:"debug"`
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
		_ = DB.Close()
	}
	if c.Username == "" {
		return errors.New("MySQL conf incomplete")
	}
	if c.Dbname == "" {
		c.Dbname = c.Username
	}
	if c.Address == "" {
		c.Address = "127.0.0.1"
	}
	if strings.Index(c.Address, ":") == -1 {
		c.Address += ":3306"
	}
	db, err := gorm.Open(
		"mysql",
		fmt.Sprintf("%s:%s@(%s)/%s?charset=%s&parseTime=True&loc=Local", c.Username, c.Password, c.Address, c.Dbname, c.Charset),
	)
	if err != nil {
		return fmt.Errorf("fail connect MySQL: %v", err)
	}

	db.DB().SetMaxIdleConns(int(c.MaxIdle))
	db.DB().SetMaxOpenConns(int(c.MaxOpen))
	db.SingularTable(true)

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return c.Prefix + defaultTableName
	}

	dbNormal = db
	dbDebug = db.Debug()
	if c.Debug {
		DB = dbDebug
	} else {
		DB = dbNormal
	}
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
		c.Address = "127.0.0.1"
	}
	if strings.Index(c.Address, ":") == -1 {
		c.Address += ":6379"
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
