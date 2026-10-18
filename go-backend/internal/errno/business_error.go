package errno

// BusinessError 自定义一个业务异常，和内置的异常类型区分开，便于定制化输出错误信息
type BusinessError struct {
	Code    int
	Message string
}

func (e *BusinessError) Error() string { // 实现 error 接口
	return e.Message
}

func New(code ErrorCode) *BusinessError {
	return &BusinessError{
		Code:    code.Code,
		Message: code.Message,
	}
}

func NewWithMessage(code ErrorCode, message string) *BusinessError {
	return &BusinessError{
		Code:    code.Code,
		Message: message,
	}
}

// AsBusinessError 尝试将 error 转换为 BusinessError
// 类似 Java 的 instanceof 判断
func AsBusinessError(err error) (*BusinessError, bool) {
	if err == nil {
		return nil, false
	}
	bizErr, ok := err.(*BusinessError)
	return bizErr, ok
}
