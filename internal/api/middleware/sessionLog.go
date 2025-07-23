package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"NewsEyeTracking/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
		
		// 记录请求开始信息
		logger.Logger.Info("开始处理Session请求",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Time("start_time", startTime),
		)
		
		// 记录重要的请求头
		if auth := c.GetHeader("Authorization"); auth != "" {
			logger.Logger.Debug("请求包含Authorization头",
				zap.String("path", path),
				zap.String("auth_preview", truncateString(auth, 20)+"...(已截断)"),
			)
		}
		if contentType := c.GetHeader("Content-Type"); contentType != "" {
			logger.Logger.Debug("请求Content-Type",
				zap.String("path", path),
				zap.String("content_type", contentType),
			)
		}

		// 读取并记录请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 重新设置请求体，以便后续中间件和处理程序能够读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		if len(requestBody) > 0 {
			logger.Logger.Debug("收到请求体",
				zap.String("path", path),
				zap.Int("body_size", len(requestBody)),
			)
			// 尝试解析JSON请求体
			if isJSONContent(c.GetHeader("Content-Type")) {
				var jsonData interface{}
				if err := json.Unmarshal(requestBody, &jsonData); err == nil {
					// 格式化输出JSON
					formattedJSON, _ := json.MarshalIndent(jsonData, "", "  ")
					logger.Logger.Debug("请求体JSON内容",
						zap.String("path", path),
						zap.String("json_content", string(formattedJSON)),
					)
				} else {
					logger.Logger.Debug("请求体内容(JSON解析失败)",
						zap.String("path", path),
						zap.String("raw_content", truncateString(string(requestBody), 500)),
					)
				}
			} else {
				logger.Logger.Debug("请求体内容",
					zap.String("path", path),
					zap.String("content", truncateString(string(requestBody), 500)),
				)
			}
		} else {
			logger.Logger.Debug("无请求体", zap.String("path", path))
		}

		// 记录URL参数
		if params := c.Request.URL.Query(); len(params) > 0 {
			logger.Logger.Debug("URL参数",
				zap.String("path", path),
				zap.Any("params", params),
			)
		}

		// 记录路径参数
		if sessionID := c.Param("session_id"); sessionID != "" {
			logger.Logger.Debug("Session ID",
				zap.String("path", path),
				zap.String("session_id", sessionID),
			)
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

		fields := []zap.Field{
			zap.Int("status_code", statusCode),
			zap.Duration("latency", latency),
			zap.String("path", path),
		}

		// 如果是404错误，给出特别的说明
		if statusCode == 404 {
			logger.Logger.Warn("404错误分析",
				append(fields, zap.String("note", "检测到错误的路径"))...,
			)
			if strings.Contains(path, "/api/v1/session/") {
				logger.Logger.Warn("路径修正建议",
					zap.String("wrong_path", path),
					zap.String("correct_path", strings.Replace(path, "/api/v1/session/", "/api/v1/sessions/", 1)),
				)
				logger.Logger.Warn("前端需要修正路径",
					zap.String("current", "session"),
					zap.String("suggested", "sessions (复数)"),
				)
			} else {
				logger.Logger.Warn("Session路径不存在",
					zap.String("path", path),
				)
			}
		}

// 记录响应体
		responseBody := responseWriter.body.String()
		if responseBody != "" {
			logger.Logger.Debug("响应体内容",
				zap.String("path", path),
				zap.Int("response_size", len(responseBody)),
			)
			// 尝试解析响应JSON
			var jsonResponse interface{}
			if err := json.Unmarshal([]byte(responseBody), &jsonResponse); err == nil {
				formattedResponse, _ := json.MarshalIndent(jsonResponse, "", "  ")
				logger.Logger.Debug("响应JSON内容",
					zap.String("json_content", string(formattedResponse)),
				)
			} else {
				logger.Logger.Debug("响应内容(非JSON)",
					zap.String("raw_content", truncateString(responseBody, 500)),
				)
			}
		} else {
			logger.Logger.Debug("无响应体", zap.String("path", path))
		}

		// 记录错误信息
		if len(c.Errors) > 0 {
			logger.Logger.Error("错误信息",
				zap.String("path", path),
				zap.String("errors", c.Errors.String()),
			)
		}

		// 记录从上下文中获取的用户信息
		if userID, exists := c.Get("userID"); exists {
			logger.Logger.Debug("用户ID",
				zap.Any("user_id", userID),
			)
		}

		// 记录完成日志
		logger.Logger.Info("Session请求处理完成", append(fields,
			zap.String("method", method),
			zap.String("client_ip", clientIP),
		)...)
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
