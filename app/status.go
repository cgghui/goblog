package app

// 响应到浏览器的状态大全
const (
	StatusOK = 0 // 正确的 正常的 成功的

	StatusForbidden = 10000 // http status 403 表示请求遭到系统拒绝
	StatusNotFound  = 10001 // http status 404 表示请求的资源不存在

	StatusQueryInvalid = 11000 // 传入的参数无效
	StatusUserNotExist = 11001
	StatusUserLocked   = 11002
	StatusCaptchaError = 12000
	StatusPasswordErr  = 11003
)

var statusRet = map[int]bool{
	StatusOK:           true,
	StatusForbidden:    false,
	StatusNotFound:     false,
	StatusQueryInvalid: false,
	StatusUserNotExist: false,
	StatusUserLocked:   false,
	StatusCaptchaError: false,
	StatusPasswordErr:  false,
}

var statusMsg = map[int]string{
	StatusOK:           "success",
	StatusForbidden:    "forbidden",
	StatusNotFound:     "not found",
	StatusQueryInvalid: "无效参数",
	StatusUserNotExist: "账号不存在",
	StatusUserLocked:   "账号被锁定",
	StatusCaptchaError: "验证码错误",
	StatusPasswordErr:  "密码错误",
}

// StatusRet 返回相关结果
func StatusRet(code int) (bool, string) {
	return statusRet[code], statusMsg[code]
}
