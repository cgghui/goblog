package app

// 响应到浏览器的状态大全
const (
	StatusOK        = 0     // 正确的 正常的 成功的
	StatusForbidden = 10000 // http status 403 表示请求遭到系统拒绝
	StatusNotFound  = 10001 // http status 404 表示请求的资源不存在
)

var statusRet = map[int]bool{
	StatusOK:        true,
	StatusForbidden: false,
	StatusNotFound:  false,
}

var statusMsg = map[int]string{
	StatusOK:        "success",
	StatusForbidden: "forbidden",
	StatusNotFound:  "not found",
}

// StatusRet 返回相关结果
func StatusRet(code int) (bool, string) {
	return statusRet[code], statusMsg[code]
}
