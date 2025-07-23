package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"NewsEyeTracking/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DevLogger 开发模式专用的详细日志中间件
func DevLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只在开发模式或启用详细日志时使用
		if !isDevelopmentMode() {
			c.Next()
			return
		}

		// 跳过眼动数据相关的路径，避免记录大量眼动参数
		if isEyeTrackingPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		// 记录请求开始
		logger.Logger.Info("🚀 开始处理请求",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Time("start_time", startTime),
		)

		// 记录所有请求头
		logger.Logger.Debug("📋 请求头信息",
			zap.String("path", path),
			zap.Any("headers", c.Request.Header),
		)

		// 记录URL参数
		if params := c.Request.URL.Query(); len(params) > 0 {
			logger.Logger.Debug("🔍 URL参数",
				zap.String("path", path),
				zap.Any("query_params", params),
			)
		}

		// 记录路径参数
		if ginParams := c.Params; len(ginParams) > 0 {
			paramMap := make(map[string]string)
			for _, param := range ginParams {
				paramMap[param.Key] = param.Value
			}
			logger.Logger.Debug("🎯 路径参数",
				zap.String("path", path),
				zap.Any("path_params", paramMap),
			)
		}

		// 读取并记录请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		if len(requestBody) > 0 {
			logger.Logger.Debug("📦 收到请求体",
				zap.String("path", path),
				zap.Int("body_size", len(requestBody)),
				zap.String("content_type", c.GetHeader("Content-Type")),
			)

			// 尝试解析JSON请求体
			if isJSONContent(c.GetHeader("Content-Type")) {
				var jsonData interface{}
				if err := json.Unmarshal(requestBody, &jsonData); err == nil {
					formattedJSON, _ := json.MarshalIndent(jsonData, "", "  ")
					logger.Logger.Debug("📄 请求体JSON内容",
						zap.String("path", path),
						zap.String("json_content", string(formattedJSON)),
					)
				} else {
					logger.Logger.Debug("❌ JSON解析失败，显示原始内容",
						zap.String("path", path),
						zap.String("raw_content", truncateString(string(requestBody), 1000)),
						zap.String("parse_error", err.Error()),
					)
				}
			} else {
				logger.Logger.Debug("📝 请求体内容（非JSON）",
					zap.String("path", path),
					zap.String("content", truncateString(string(requestBody), 1000)),
				)
			}
		} else {
			logger.Logger.Debug("📭 无请求体", zap.String("path", path))
		}

		// 创建响应写入器来捕获响应
		responseWriter := &responseBodyWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = responseWriter

		// 处理请求
		c.Next()

		// 记录响应信息
		endTime := time.Now()
		latency := endTime.Sub(startTime)
		statusCode := c.Writer.Status()

		// 记录响应体
		responseBody := responseWriter.body.String()
		if responseBody != "" {
			logger.Logger.Debug("📤 响应体信息",
				zap.String("path", path),
				zap.Int("response_size", len(responseBody)),
				zap.Int("status_code", statusCode),
			)

			// 尝试解析响应JSON
			var jsonResponse interface{}
			if err := json.Unmarshal([]byte(responseBody), &jsonResponse); err == nil {
				formattedResponse, _ := json.MarshalIndent(jsonResponse, "", "  ")
				logger.Logger.Debug("📋 响应JSON内容",
					zap.String("path", path),
					zap.String("json_content", string(formattedResponse)),
				)
			} else {
				logger.Logger.Debug("📝 响应内容（非JSON）",
					zap.String("path", path),
					zap.String("raw_content", truncateString(responseBody, 1000)),
				)
			}
		} else {
			logger.Logger.Debug("📭 无响应体", zap.String("path", path))
		}

		// 记录错误信息
		if len(c.Errors) > 0 {
			logger.Logger.Error("❌ 处理过程中出现错误",
				zap.String("path", path),
				zap.String("errors", c.Errors.String()),
			)
		}

		// 记录从上下文中获取的用户信息
		if userID, exists := c.Get("userID"); exists {
			logger.Logger.Debug("👤 用户信息",
				zap.String("path", path),
				zap.Any("user_id", userID),
			)
		}

		// 构建完成日志字段
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status_code", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.Time("end_time", endTime),
		}

		// 根据状态码和延迟选择合适的日志级别和emoji
		var emoji string
		switch {
		case statusCode >= 500:
			emoji = "💥"
			logger.Logger.Error(emoji+" 服务器错误", fields...)
		case statusCode >= 400:
			emoji = "⚠️"
			logger.Logger.Warn(emoji+" 客户端错误", fields...)
		case statusCode >= 300:
			emoji = "↩️"
			logger.Logger.Info(emoji+" 重定向", fields...)
		case latency > 1*time.Second:
			emoji = "🐌"
			logger.Logger.Warn(emoji+" 请求处理较慢", fields...)
		case latency > 500*time.Millisecond:
			emoji = "⏳"
			logger.Logger.Info(emoji+" 请求处理完成", fields...)
		default:
			emoji = "✅"
			logger.Logger.Info(emoji+" 请求处理完成", fields...)
		}

		// 添加性能分析信息
		if latency > 100*time.Millisecond {
			logger.Logger.Info("⏱️ 性能分析",
				zap.String("path", path),
				zap.Duration("latency", latency),
				zap.String("performance_note", getPerformanceNote(latency)),
			)
		}
	}
}

// isDevelopmentMode 检查是否为开发模式
func isDevelopmentMode() bool {
	env := os.Getenv("GIN_MODE")
	logLevel := os.Getenv("LOG_LEVEL")
	return env != "release" || logLevel == "debug"
}

// getPerformanceNote 根据延迟返回性能建议
func getPerformanceNote(latency time.Duration) string {
	switch {
	case latency > 5*time.Second:
		return "请求非常慢，需要优化"
	case latency > 2*time.Second:
		return "请求较慢，建议检查"
	case latency > 1*time.Second:
		return "请求稍慢，可以优化"
	case latency > 500*time.Millisecond:
		return "响应时间可接受"
	default:
		return "响应快速"
	}
}

// isEyeTrackingPath 检查是否为眼动数据相关的路径
func isEyeTrackingPath(path string) bool {
	// 定义需要跳过记录的眼动数据路径
	eyeTrackingPaths := []string{
		"/api/v1/sessions/",  // sessions 相关的数据上传路径
	}

	for _, eyePath := range eyeTrackingPaths {
		if strings.Contains(path, eyePath) && strings.HasSuffix(path, "/data") {
			return true // 只跳过 /sessions/:session_id/data 路径
		}
	}
	return false
}
