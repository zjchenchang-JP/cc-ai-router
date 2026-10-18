package common

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// BaseResponse 封装统一响应结构
// struct tag（json:"code"）控制 JSON 序列化
type BaseResponse struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

// any 是 Go 1.18 引入的类型别名，等价于 interface{}，类似 Java 的泛型 T
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, BaseResponse{
		Code:    0,
		Data:    data,
		Message: "ok",
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, BaseResponse{
		Code:    code,
		Data:    nil,
		Message: message,
	})
}
