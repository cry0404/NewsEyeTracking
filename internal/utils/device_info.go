package utils

import (
	"NewsEyeTracking/internal/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ExtractDeviceInfoFromHeaders 从请求头中提取设备信息
func ExtractDeviceInfoFromHeaders(c *gin.Context) *models.DeviceInfo {
	deviceInfo := &models.DeviceInfo{
		UserAgent: c.GetHeader("User-Agent"),
	}

	// 从自定义请求头中获取屏幕信息（需要前端设置）
	if screenWidth := c.GetHeader("X-Screen-Width"); screenWidth != "" {
		if width, err := strconv.Atoi(screenWidth); err == nil {
			deviceInfo.ScreenWidth = width
		}
	}

	if screenHeight := c.GetHeader("X-Screen-Height"); screenHeight != "" {
		if height, err := strconv.Atoi(screenHeight); err == nil {
			deviceInfo.ScreenHeight = height
		}
	}

	if viewportWidth := c.GetHeader("X-Viewport-Width"); viewportWidth != "" {
		if width, err := strconv.Atoi(viewportWidth); err == nil {
			deviceInfo.ViewportWidth = width
		}
	}

	if viewportHeight := c.GetHeader("X-Viewport-Height"); viewportHeight != "" {
		if height, err := strconv.Atoi(viewportHeight); err == nil {
			deviceInfo.ViewportHeight = height
		}
	}

	return deviceInfo
}

// ExtractDeviceInfoFromRequest 从请求体中提取设备信息（备用方法）
// 当前端无法设置自定义请求头时使用
func ExtractDeviceInfoFromRequest(c *gin.Context, reqDeviceInfo *models.DeviceInfo) *models.DeviceInfo {
	deviceInfo := &models.DeviceInfo{
		UserAgent: c.GetHeader("User-Agent"),
	}

	// 如果请求体中包含设备信息，则使用它
	if reqDeviceInfo != nil {
		if reqDeviceInfo.ScreenWidth > 0 {
			deviceInfo.ScreenWidth = reqDeviceInfo.ScreenWidth
		}
		if reqDeviceInfo.ScreenHeight > 0 {
			deviceInfo.ScreenHeight = reqDeviceInfo.ScreenHeight
		}
		if reqDeviceInfo.ViewportWidth > 0 {
			deviceInfo.ViewportWidth = reqDeviceInfo.ViewportWidth
		}
		if reqDeviceInfo.ViewportHeight > 0 {
			deviceInfo.ViewportHeight = reqDeviceInfo.ViewportHeight
		}
	}

	// 回退到请求头获取
	if deviceInfo.ScreenWidth == 0 {
		if screenWidth := c.GetHeader("X-Screen-Width"); screenWidth != "" {
			if width, err := strconv.Atoi(screenWidth); err == nil {
				deviceInfo.ScreenWidth = width
			}
		}
	}

	if deviceInfo.ScreenHeight == 0 {
		if screenHeight := c.GetHeader("X-Screen-Height"); screenHeight != "" {
			if height, err := strconv.Atoi(screenHeight); err == nil {
				deviceInfo.ScreenHeight = height
			}
		}
	}

	if deviceInfo.ViewportWidth == 0 {
		if viewportWidth := c.GetHeader("X-Viewport-Width"); viewportWidth != "" {
			if width, err := strconv.Atoi(viewportWidth); err == nil {
				deviceInfo.ViewportWidth = width
			}
		}
	}

	if deviceInfo.ViewportHeight == 0 {
		if viewportHeight := c.GetHeader("X-Viewport-Height"); viewportHeight != "" {
			if height, err := strconv.Atoi(viewportHeight); err == nil {
				deviceInfo.ViewportHeight = height
			}
		}
	}

	return deviceInfo
}

// ParseUserAgent 解析 User-Agent 获取浏览器和操作系统信息
func ParseUserAgent(userAgent string) (browser, os string) {
	userAgent = strings.ToLower(userAgent)
	
	// 检测浏览器
	if strings.Contains(userAgent, "chrome") {
		browser = "Chrome"
	} else if strings.Contains(userAgent, "firefox") {
		browser = "Firefox"
	} else if strings.Contains(userAgent, "safari") && !strings.Contains(userAgent, "chrome") {
		browser = "Safari"
	} else if strings.Contains(userAgent, "edge") {
		browser = "Edge"
	} else {
		browser = "Unknown"
	}

	// 检测操作系统
	if strings.Contains(userAgent, "windows") {
		os = "Windows"
	} else if strings.Contains(userAgent, "macintosh") || strings.Contains(userAgent, "mac os") {
		os = "macOS"
	} else if strings.Contains(userAgent, "linux") {
		os = "Linux"
	} else if strings.Contains(userAgent, "android") {
		os = "Android"
	} else if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad") {
		os = "iOS"
	} else {
		os = "Unknown"
	}

	return browser, os
}
