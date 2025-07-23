

package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"NewsEyeTracking/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"github.com/mattn/go-colorable"
	"github.com/fatih/color"
)

// Logger zap日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		// 颜色配置
		color.NoColor = false // 启用彩色输出
		colorOutput := colorable.NewColorableStdout()
		methodColor := color.New(color.FgYellow).SprintFunc()
		getColor := color.New(color.FgCyan).SprintFunc()
		postColor := color.New(color.FgGreen).SprintFunc()

		// 确定方法颜色
		coloredMethod := method
		switch method {
		case "GET":
			coloredMethod = getColor(method)
		case "POST":
			coloredMethod = postColor(method)
		default:
			coloredMethod = methodColor(method)
		}

		// 记录请求体
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

		// 输出请求信息（带色彩）
		fmt.Fprintf(colorOutput, "→ %s %s from %s\n", coloredMethod, path, clientIP)

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

		// 过滤眼动信息路径的JSON输出，但保留状态码日志
		if !isEyeTrackingDataPath(path) {
			// 输出带颜色的响应信息
			statusColor := getStatusColor(statusCode)
			fmt.Fprintf(colorOutput, "← %s %s %s [%s]\n", 
				coloredMethod, path, statusColor(fmt.Sprintf("%d", statusCode)), latency)
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


// isEyeTrackingDataPath 检查是否为需要跳过的眼动数据路径
func isEyeTrackingDataPath(path string) bool {
	// 定义需要跳过记录详细JSON的路径
	return strings.Contains(path, "/sessions/") && strings.HasSuffix(path, "/data")
}

func getStatusColor(statusCode int) func(a ...interface{}) string {
	switch {
	case statusCode >= 500:
		return color.New(color.FgRed).SprintFunc()
	case statusCode >= 400:
		return color.New(color.FgYellow).SprintFunc()
	case statusCode >= 300:
		return color.New(color.FgBlue).SprintFunc()
	default:
		return color.New(color.FgGreen).SprintFunc()
	}
}

