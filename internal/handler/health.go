package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/rainlf/mango-crew/pkg/response"
)

// RegisterHealthRoutes 注册健康检查路由
func RegisterHealthRoutes(r *gin.RouterGroup) {
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, "ok")
	})
}
