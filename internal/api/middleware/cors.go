package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS 跨域资源共享中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置允许的来源
		c.Header("Access-Control-Allow-Origin", "*")

		// 设置允许的HTTP方法
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// 设置允许的请求头
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// 设置允许凭证
		c.Header("Access-Control-Allow-Credentials", "true")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
