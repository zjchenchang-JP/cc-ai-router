package errno

// ErrorCode 自定义错误码，对错误进行收敛，便于前端统一处理
// 自定义错误码时，建议跟主流的错误码（比如 HTTP 错误码）的含义保持一致，比如 "未登录" 定义为 40100，和 HTTP 401 错误（用户需要进行身份认证）保持一致，会更容易理解。
// 错误码不要完全连续，预留一些间隔，便于后续扩展。
type ErrorCode struct {
	Code    int
	Message string
}

// Go 没有枚举类型，但可以用 var 声明包级别的变量来模拟枚举。这些变量在程序启动时初始化，全局可用
var (
	Success        = ErrorCode{Code: 0, Message: "ok"}
	ParamsError    = ErrorCode{Code: 40000, Message: "请求参数错误"}
	NotLoginError  = ErrorCode{Code: 40100, Message: "未登录"}
	NoAuthError    = ErrorCode{Code: 40101, Message: "无权限"}
	TooManyRequest = ErrorCode{Code: 42900, Message: "请求过于频繁"}
	NotFoundError  = ErrorCode{Code: 40400, Message: "请求数据不存在"}
	ForbiddenError = ErrorCode{Code: 40300, Message: "禁止访问"}
	SystemError    = ErrorCode{Code: 50000, Message: "系统内部异常"}
	OperationError = ErrorCode{Code: 50001, Message: "操作失败"}
)
