package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rainlf/mango-crew/pkg/logger"
	"github.com/rainlf/mango-crew/pkg/response"
)

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				errText := fmt.Sprint(err)
				c.Error(fmt.Errorf("panic recovered: %s", errText))

				logger.Error("panic recovered",
					logger.String("error", errText),
					logger.String("path", c.Request.URL.Path),
				)
				response.InternalError(c, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}
