package middleware

import (
	"bytes"
	"strings"

	"github.com/gin-gonic/gin"
)

// isJSONContent 检查是否为JSON内容类型
func isJSONContent(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "application/json")
}

// truncateString 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// responseBodyWriter 用于捕获响应体的写入器
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
