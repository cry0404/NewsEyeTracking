package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 自定义日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求 - 调用下一个中间件或处理函数
		c.Next()

		endTime := time.Now()
		latency := endTime.Sub(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		errorMessage := ""
		if len(c.Errors) > 0 {
			errorMessage = c.Errors.String()
		}

		log.Printf("[GIN] %s | %3d | %13v | %15s | %-7s | %s | %s",
			endTime.Format("2006/01/02 - 15:04:05"), // 修复时间格式
			statusCode,
			latency,
			clientIP,
			method,
			path,
			errorMessage,
		)

		if statusCode >= 500 {
			log.Printf("[ERROR] 服务器错误：%v", c.Errors.String())
		}
	}
}
