package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health 健康检查
func (h *HealthHandler) Health(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

// RegisterHealthRoutes 注册健康检查路由
func RegisterHealthRoutes(r *gin.RouterGroup) {
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
}
