package middleware

import (
	"NewsEyeTracking/internal/models"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// CORS 跨域资源共享中间件，根据环境变量配置不同的CORS策略
func CORS() gin.HandlerFunc {
	// 尝试加载环境变量，如果失败只记录日志，不中断程序，也可以考虑中断
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	env := os.Getenv("CURRENT_ENV")
	// 如果环境变量未设置，默认为开发环境
	if env == "" {
		env = "dev"
	}

	switch env {
	case "dev":
		return func(c *gin.Context) {
			// 开发环境：允许所有来源
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")

			// 处理预检请求
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}

			c.Next()
		}
	case "release":
		return func(c *gin.Context) {
			// 生产环境：更严格的CORS配置
			allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
			if allowedOrigins == "" {
				// 如果没有配置允许的域名，返回错误
				c.JSON(http.StatusInternalServerError, models.ErrorResponse(
					models.ErrorCodeInternalError,
					"CORS配置错误",
					"Production environment requires ALLOWED_ORIGINS configuration",
				))
				c.Abort()
				return
			}

			c.Header("Access-Control-Allow-Origin", allowedOrigins)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")

			// 处理预检请求
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}

			c.Next()
		}
	default:
		return func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrorCodeInternalError,
				"环境配置错误",
				"Unknown environment configuration",
			))
			c.Abort()
		}
	}
}
