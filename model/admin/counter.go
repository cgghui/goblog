package admin

import (
	"encoding/json"
	"goblog/app"
	"goblog/model/config"
	"time"
)

// CounterField 统计字段类型
type CounterField string

// 管理员登录时的错误记录字段
var (
	CounterPassword CounterField = "password"
	CounterCaptcha  CounterField = "captcha"
)

// CounterCheck 检查n是否大于field所记录的数，大于则返回true 否则返回false
func (a *Admins) CounterCheck(field CounterField, n int) bool {
	num, _ := a.CounterGet(field)
	return num > n
}

// CounterGet 取出错误记录
func (a *Admins) CounterGet(field CounterField) (int, *time.Time) {

	data, err := app.RedisConn.HGet(a.eckey(), string(field)).Result()
	if err != nil {
		return 0, nil
	}

	ret := make([]interface{}, 0)
	if err := json.Unmarshal([]byte(data), &ret); err != nil || len(ret) != 3 {
		a.CounterClear(field)
		return 0, nil
	}

	if time.Unix(int64(ret[1].(float64)), 0).Sub(time.Now()).Seconds() <= 0 {
		a.CounterClear(field)
		return 0, nil
	}

	t := time.Unix(int64(ret[2].(float64)), 0)

	return int(ret[0].(float64)), &t
}

// CounterIncr 增加一次错误记录
func (a *Admins) CounterIncr(field CounterField) {
	num, _ := a.CounterGet(field)
	num++
	data, err := json.Marshal([]interface{}{
		num,
		config.GetConfigField("admin", "login_counter_expire").TimeNowAddToUnix(),
		time.Now().Unix(),
	})
	if err != nil {
		panic(err)
	}
	if err := app.RedisConn.HSet(a.eckey(), string(field), data).Err(); err != nil {
		panic(err)
	}
	return
}

// CounterClear 清除错误记录
func (a *Admins) CounterClear(fields ...CounterField) {

	if len(fields) == 0 {
		if err := app.RedisConn.Del(a.eckey()).Err(); err != nil {
			panic(err)
		}
		return
	}

	fs := make([]string, 0)
	for i := 0; i < len(fields); i++ {
		if len(fields[i]) == 0 {
			continue
		}
		fs = append(fs, string(fields[i]))
	}
	if len(fs) == 0 {
		return
	}
	if err := app.RedisConn.HDel(a.eckey(), fs...).Err(); err != nil {
		panic(err)
	}
	return
}

func (a *Admins) eckey() string {
	return "AdminsCounter_" + a.Username
}
