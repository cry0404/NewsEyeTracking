package middleware

import (
	"log"
	"net/http"

	"NewsEyeTracking/internal/models"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 全局错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v", err)

				if !c.Writer.Written() {
					c.JSON(http.StatusInternalServerError, models.ErrorResponse(
						models.ErrorCodeInternalError,
						"服务器内部错误",
						"请稍后重试",
					))
				}
			}
		}()

		c.Next()

		// 处理在处理程序中设置的错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Printf("[ERROR] %v", err.Error())

			// 如果响应还没有写入，返回错误响应
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse(
					models.ErrorCodeInternalError,
					"请求处理失败",
					err.Error(),
				))
			}
		}
	}
}
