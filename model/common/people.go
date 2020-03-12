package common

import "regexp"

// Checker Check interface
type Checker interface {
	Check() bool
	Name() string
	String() string
}

// Gender 性别
type Gender string

// Check 是否是合法的性别
func (g Gender) Check() bool {
	for _, gender := range Genders {
		if gender == g {
			return true
		}
	}
	return false
}

// Name 字段名
func (g Gender) Name() string {
	return "gender"
}

// String 转字符串
func (g Gender) String() string {
	return string(g)
}

// Genders 性别列表
var Genders = []Gender{"M", "W", "X"}

// Mobile 手机号码类型
type Mobile string

// Check 是否为手机号码 true正确的手机号码
func (m Mobile) Check() bool {
	return RegexpChinaMobile.MatchString(string(m))
}

// Name 字段名
func (m Mobile) Name() string {
	return "mobile"
}

// String 转字符串
func (m Mobile) String() string {
	return string(m)
}

// RegexpChinaMobile  验证手机号码的表达式
var RegexpChinaMobile = regexp.MustCompile(`^1([38][0-9]|14[579]|5[^4]|16[6]|7[1-35-8]|9[189])\d{8}$`)

// Email 邮箱地址类型
type Email string

// Check 是否为邮箱地址 true正确的邮箱地址
func (e Email) Check() bool {
	return RegexpEmail.MatchString(string(e))
}

// Name 字段名
func (e Email) Name() string {
	return "email"
}

// String 转字符串
func (e Email) String() string {
	return string(e)
}

// RegexpEmail 验证邮箱地址的表达式
var RegexpEmail = regexp.MustCompile(`\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`)
