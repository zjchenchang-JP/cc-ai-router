package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/zjchenchang-JP/cc-ai-router/go-backend/internal/common"
	"github.com/zjchenchang-JP/cc-ai-router/go-backend/internal/errno"
	"log"
)

// Recovery 全局恢复中间件
// Go 中没有全局异常处理器，但可以通过 recover 机制捕获 panic，防止程序崩溃
// 等同于 Java 的 @RestControllerAdvice 全局异常处理器，能够捕获所有未处理的 panic，返回统一的错误响应
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic recovered: %v", recovered)
				common.Error(c, errno.SystemError.Code, "系统错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}
