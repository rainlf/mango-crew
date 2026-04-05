package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rainlf/mgtt-go/pkg/logger"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		fields := []logger.Field{
			logger.Int("status", statusCode),
			logger.String("method", method),
			logger.String("path", path),
			logger.String("ip", clientIP),
			logger.Duration("latency", latency),
		}

		if len(c.Errors) > 0 {
			logger.Error("request failed", append(fields, logger.String("error", c.Errors.String()))...)
		} else {
			logger.Info("request completed", fields...)
		}
	}
}
