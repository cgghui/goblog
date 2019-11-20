package model

import (
	"encoding/json"
	"fmt"
	"goblog/app"
	"strconv"
)

//conf 记录所有的配置数据
var conf map[string]map[string]*Configs

// Configs 全局参数配置
type Configs struct {
	ID        uint   `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	Namespace string `gorm:"type:varchar(32)"`
	Field     string `gorm:"type:varchar(64)"`
	Type      string `gorm:"type:enum('string','int','float','json')"`
	Value     string `gorm:"type:varchar(512)"`
}

// ConfigAdminLCC 管理员登录时启用验证码的条件
// Config Admin Login Captcha Condition
type ConfigAdminLCC struct {
	Password       int `json:"pwd_errn"`
	Captcha        int `json:"captcha_errn"`
	GoogleAuthCode int `json:"google_authcode_errn"`
}

func init() {
	cfgs := []Configs{}
	app.DBConn.Find(&cfgs)
	conf = make(map[string]map[string]*Configs)
	for i, cfg := range cfgs {
		if _, ok := conf[cfg.Namespace]; !ok {
			conf[cfg.Namespace] = make(map[string]*Configs)
		}
		conf[cfg.Namespace][cfg.Field] = &cfgs[i]
	}
}

// GetConfigField 取出一个值
func GetConfigField(namespace, field string) *Configs {
	if val, ok := conf[namespace][field]; ok {
		return val
	}
	panic(fmt.Sprintf("Error: conf[%s][%s] does not exist", namespace, field))
}

// GetConfigNamespace 取出一个空间的配置
func GetConfigNamespace(namespace string) map[string]*Configs {
	if val, ok := conf[namespace]; ok {
		return val
	}
	panic(fmt.Sprintf("Error: conf[%s] does not exist", namespace))
}

// String value为string
func (c *Configs) String() string {
	if c.Type != "string" {
		panic(fmt.Sprintf("Error: %s[%s] = %s value not string", c.Namespace, c.Field, c.Value))
	}
	return c.Value
}

// Int value为int
func (c *Configs) Int() int {
	if c.Type != "int" {
		panic(fmt.Sprintf("Error: %s[%s] = %s value not int", c.Namespace, c.Field, c.Value))
	}
	ret, err := strconv.Atoi(c.Value)
	if err != nil {
		panic(fmt.Sprintf("Error: %s[%s] = %s value not int %v", c.Namespace, c.Field, c.Value, err))
	}
	return ret
}

// BindJSON BIND JSON
func (c *Configs) BindJSON(result interface{}) {
	if err := json.Unmarshal([]byte(c.Value), &result); err != nil {
		panic(fmt.Sprintf("Error: %s[%s] = %s BindJSON %v", c.Namespace, c.Field, c.Value, err))
	}
	return
}

// Val 获取结果
func (c *Configs) Val() interface{} {
	switch c.Type {
	case "int":
		{
			ret, err := strconv.Atoi(c.Value)
			if err != nil {
				panic(fmt.Sprintf("Error: %s[%s] = %s value not int %v", c.Namespace, c.Field, c.Value, err))
			}
			return ret
		}
	case "float":
		{
			ret, err := strconv.ParseFloat(c.Value, 64)
			if err != nil {
				panic(fmt.Sprintf("Error: %s[%s] = %s value not float ( 64 or 32 ) %v", c.Namespace, c.Field, c.Value, err))
			}
			return ret
		}
	case "json":
		{
			ret := make(map[string]interface{})
			if err := json.Unmarshal([]byte(c.Value), &ret); err != nil {
				panic(fmt.Sprintf("Error: %s[%s] = %s value not json %v", c.Namespace, c.Field, c.Value, err))
			}
			return ret
		}
	}
	return c.Value
}
