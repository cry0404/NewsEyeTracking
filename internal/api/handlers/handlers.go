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

// 一个架构设计，通过 handler 来处理所有定义的服务
func NewHandlers(services *service.Services) *Handlers {
	return &Handlers{services: services}
}


func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}


func (h *Handlers) Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": "1.0.0"})
}
