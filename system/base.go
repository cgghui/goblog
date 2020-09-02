package system

import "errors"

const (
	Name    = "GoBlog"               // 系统名称
	Version = ".gitignore.0.0"       // 系统版本
	HomeURL = "http://www.04559.com" // 系统官方网址
	Author  = "ChenGuangHui"         // 系统作者
)

// ConfMySQL MySQL配置参数
type ConfMySQL struct {
	Enable   bool   `json:"enable"`
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
	Dbname   string `json:"dbname"`
	Charset  string `json:"charset"`
	MaxIdle  uint16 `json:"max_idle"`
	MaxOpen  uint16 `json:"max_open"`
}

// ConnectMySQL 连接MySQL 成功返回nil
func ConnectMySQL(c *ConfMySQL) error {
	if !c.Enable {
		return nil
	}
	if c.Dbname == "" {
		c.Dbname = c.Username
	}
	if c.Address == "" || c.Username == "" {
		return errors.New("MySQL conf incomplete")
	}

	return nil
}
