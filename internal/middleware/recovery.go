package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rainlf/mgtt-go/pkg/logger"
	"github.com/rainlf/mgtt-go/pkg/response"
)

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered",
					logger.String("error", err.(error).Error()),
					logger.String("path", c.Request.URL.Path),
				)
				response.InternalError(c, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}
