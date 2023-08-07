package errcode

// 框架组件错误码
var (
	LibNoRoute        = &ErrCode{1001, "路由未找到", ""}
	LibRateLimit      = &ErrCode{1002, "服务繁忙，请稍后再试", ""}
	LibNotInWhitelist = &ErrCode{1003, "您无权限访问!请使用浏览器打开`ipip.net`网站,获取ip后,联系管理员", ""}
)
