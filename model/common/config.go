package common

import (
	"encoding/json"
	"fmt"
	"goblog/app"
	"strconv"
	"time"
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

// Get 取出一个值
func Get(namespace, field string) *Configs {
	if val, ok := conf[namespace][field]; ok {
		return val
	}
	panic(fmt.Sprintf("Error: conf[%s][%s] does not exist", namespace, field))
}

// GetNamespace 取出一个空间的配置
func GetNamespace(namespace string) map[string]*Configs {
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

// Bool value为boolean
func (c *Configs) Bool() bool {
	if c.Type != "bool" {
		panic(fmt.Sprintf("Error: %s[%s] = %s value not bool", c.Namespace, c.Field, c.Value))
	}
	return c.Value == "true"
}

// BindStruct BIND JSON
func (c *Configs) BindStruct(result interface{}) {
	if err := json.Unmarshal([]byte(c.Value), &result); err != nil {
		panic(fmt.Sprintf("Error: %s[%s] = %s BindJSON %v", c.Namespace, c.Field, c.Value, err))
	}
	return
}

// Time value为time
func (c *Configs) Time() time.Duration {
	if c.Type != "time" {
		panic(fmt.Sprintf("Error: %s[%s] = %s value not time", c.Namespace, c.Field, c.Value))
	}
	ret, err := strconv.ParseInt(c.Value, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Error: %s[%s] = %s value not time %v", c.Namespace, c.Field, c.Value, err))
	}
	return time.Duration(ret) * time.Second
}

// TimeNowAddToUnix 将时间类型的数据增加到当前时间 以时间戳返回
func (c *Configs) TimeNowAddToUnix() int64 {
	return time.Now().Add(c.Time()).Unix()
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
	case "time":
		{
			ret, err := strconv.ParseInt(c.Value, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("Error: %s[%s] = %s value not time %v", c.Namespace, c.Field, c.Value, err))
			}
			return time.Duration(ret) * time.Second
		}
	}
	return c.Value
}
