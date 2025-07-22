package middleware

import (
	"bytes"
	"encoding/json"
	
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SessionLogMiddleware 专用于记录session相关请求的中间件
func SessionLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否为session相关路径
		if !isSessionPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		
		// 记录请求头信息
		log.Printf("[SESSION_LOG] === 开始处理Session请求 ===")
		log.Printf("[SESSION_LOG] 时间: %s", startTime.Format("2006-01-02 15:04:05"))
		log.Printf("[SESSION_LOG] 方法: %s", method)
		log.Printf("[SESSION_LOG] 路径: %s", path)
		log.Printf("[SESSION_LOG] 客户端IP: %s", clientIP)
		log.Printf("[SESSION_LOG] User-Agent: %s", userAgent)
		
		// 记录重要的请求头
		if auth := c.GetHeader("Authorization"); auth != "" {
			log.Printf("[SESSION_LOG] Authorization: %s...(已截断)", truncateString(auth, 20))
		}
		if contentType := c.GetHeader("Content-Type"); contentType != "" {
			log.Printf("[SESSION_LOG] Content-Type: %s", contentType)
		}

		// 读取并记录请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 重新设置请求体，以便后续中间件和处理程序能够读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		if len(requestBody) > 0 {
			log.Printf("[SESSION_LOG] 请求体大小: %d bytes", len(requestBody))
			// 尝试解析JSON请求体
			if isJSONContent(c.GetHeader("Content-Type")) {
				var jsonData interface{}
				if err := json.Unmarshal(requestBody, &jsonData); err == nil {
					// 格式化输出JSON
					formattedJSON, _ := json.MarshalIndent(jsonData, "", "  ")
					log.Printf("[SESSION_LOG] 请求体内容:\n%s", string(formattedJSON))
				} else {
					log.Printf("[SESSION_LOG] 请求体内容 (非JSON): %s", truncateString(string(requestBody), 500))
				}
			} else {
				log.Printf("[SESSION_LOG] 请求体内容: %s", truncateString(string(requestBody), 500))
			}
		} else {
			log.Printf("[SESSION_LOG] 无请求体")
		}

		// 记录URL参数
		if params := c.Request.URL.Query(); len(params) > 0 {
			log.Printf("[SESSION_LOG] URL参数: %v", params)
		}

		// 记录路径参数
		if sessionID := c.Param("session_id"); sessionID != "" {
			log.Printf("[SESSION_LOG] Session ID: %s", sessionID)
		}

		// 创建响应写入器来捕获响应
		responseWriter := &responseBodyWriter{
			body: bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = responseWriter

		// 处理请求
		c.Next()

		// 记录响应信息
		endTime := time.Now()
		latency := endTime.Sub(startTime)
		statusCode := c.Writer.Status()

		log.Printf("[SESSION_LOG] === Session请求处理完成 ===")
		log.Printf("[SESSION_LOG] 状态码: %d", statusCode)
		log.Printf("[SESSION_LOG] 处理时间: %v", latency)
		
		// 如果是404错误，给出特别的说明
		if statusCode == 404 {
			log.Printf("[SESSION_LOG] ⚠️  404错误分析:")
			if strings.Contains(path, "/api/v1/session/") {
				log.Printf("[SESSION_LOG] ⚠️  检测到错误的路径: %s", path)
				log.Printf("[SESSION_LOG] ⚠️  正确的路径应该是: %s", strings.Replace(path, "/api/v1/session/", "/api/v1/sessions/", 1))
				log.Printf("[SESSION_LOG] ⚠️  前端需要修正路径中的 'session' 为 'sessions'（复数）")
			} else {
				log.Printf("[SESSION_LOG] ⚠️  Session路径不存在: %s", path)
			}
		}

		// 记录响应体
		responseBody := responseWriter.body.String()
		if responseBody != "" {
			log.Printf("[SESSION_LOG] 响应体大小: %d bytes", len(responseBody))
			// 尝试解析响应JSON
			var jsonResponse interface{}
			if err := json.Unmarshal([]byte(responseBody), &jsonResponse); err == nil {
				formattedResponse, _ := json.MarshalIndent(jsonResponse, "", "  ")
				log.Printf("[SESSION_LOG] 响应内容:\n%s", string(formattedResponse))
			} else {
				log.Printf("[SESSION_LOG] 响应内容: %s", truncateString(responseBody, 500))
			}
		} else {
			log.Printf("[SESSION_LOG] 无响应体")
		}

		// 记录错误信息
		if len(c.Errors) > 0 {
			log.Printf("[SESSION_LOG] 错误信息: %s", c.Errors.String())
		}

		// 记录从上下文中获取的用户信息
		if userID, exists := c.Get("userID"); exists {
			log.Printf("[SESSION_LOG] 用户ID: %v", userID)
		}

		log.Printf("[SESSION_LOG] === Session请求日志结束 ===")
		log.Printf("")
	}
}

// isSessionPath 检查是否为session相关路径
func isSessionPath(path string) bool {
	sessionPaths := []string{
		"/api/v1/sessions/",    // 正确的复数形式
		"/api/v1/session/",     // 错误的单数形式，但需要记录
		"/api/v1/heartbeat",    // 心跳相关
	}

	for _, sessionPath := range sessionPaths {
		if strings.Contains(path, sessionPath) {
			return true
		}
	}
	return false
}

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
