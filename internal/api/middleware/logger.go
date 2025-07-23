package middleware

import (
	"time"

	"NewsEyeTracking/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger zap日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		// 处理请求 - 调用下一个中间件或处理函数
		c.Next()

		endTime := time.Now()
		latency := endTime.Sub(startTime)
		statusCode := c.Writer.Status()

		// 构建日志字段
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status_code", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
		}

		// 添加错误信息（如果存在）
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// 根据状态码选择日志级别
		switch {
		case statusCode >= 500:
			logger.Logger.Error("服务器错误", fields...)
		case statusCode >= 400:
			logger.Logger.Warn("客户端错误", fields...)
		case statusCode >= 300:
			logger.Logger.Info("重定向", fields...)
		default:
			logger.Logger.Info("请求处理完成", fields...)
		}
	}
}
