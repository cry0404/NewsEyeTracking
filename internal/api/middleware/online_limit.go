package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"NewsEyeTracking/internal/models"

	"github.com/gin-gonic/gin"
)

// OnlineCounter 返回当前在线用户数
type OnlineCounter func(ctx context.Context) (int, error)

// OnlineLimit 通过计数器限制总在线人数
func OnlineLimit(maxOnline int, counter OnlineCounter) gin.HandlerFunc {
	if maxOnline <= 0 {
		// 默认不上限
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		// 给计数器一点点超时，避免请求阻塞
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()
		n, err := counter(ctx)
		if err == nil && n >= maxOnline {
			c.Header("Retry-After", "30")
			c.Header("X-Online-Users", strconv.Itoa(n))
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse(
				models.ErrorCodeTooManyRequests,
				"当前在线人数已达上限",
				"请稍后再试",
			))
			c.Abort()
			return
		}
		c.Next()
	}
}
