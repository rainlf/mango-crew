package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterHealthRoutes 注册健康检查路由
func RegisterHealthRoutes(r *gin.RouterGroup) {
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
}
