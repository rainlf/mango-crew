package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rainlf/mango-crew/pkg/response"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health 健康检查
func (h *HealthHandler) Health(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "UP",
		"service": "mango-crew",
	})
}

// RegisterHealthRoutes 注册健康检查路由
func RegisterHealthRoutes(r *gin.RouterGroup) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
			"service": "mango-crew",
		})
	})
}
