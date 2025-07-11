package handlers

import (
	"NewsEyeTracking/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handlers 处理程序结构体，持有所有依赖项
type Handlers struct {
	services *service.Services
}

// NewHandlers 创建Handlers实例
func NewHandlers(services *service.Services) *Handlers {
	return &Handlers{services: services}
}

// HealthCheck 健康检查端点
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Version 版本信息端点
func (h *Handlers) Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": "1.0.0"})
}
