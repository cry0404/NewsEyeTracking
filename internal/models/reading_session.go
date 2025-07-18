package models

import (
	"time"

	"github.com/google/uuid"
)

// DeviceInfo 设备信息， 对应
type DeviceInfo struct {
	UserAgent      string `json:"user_agent"`
	ScreenWidth    int    `json:"screen_width"`
	ScreenHeight   int    `json:"screen_height"`
	ViewportWidth  int    `json:"viewport_width"`
	ViewportHeight int    `json:"viewport_height"`
}

// ReadingSession 阅读会话模型
type ReadingSession struct {
	ID         uuid.UUID   `json:"id" db:"id"`
	UserID     uuid.UUID   `json:"user_id" db:"user_id"`
	ArticleID  int         `json:"article_id" db:"article_id"`
	StartTime  time.Time   `json:"start_time" db:"start_time"`
	EndTime    *time.Time  `json:"end_time" db:"end_time"`
	DeviceInfo *DeviceInfo `json:"device_info" db:"device_info"` // JSONB字段
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequestForArticle struct {
	ArticleID  string         `json:"article_id" binding:"required"`
	StartTime  time.Time   `json:"start_time" binding:"required"`
	DeviceInfo *DeviceInfo `json:"device_info"` //这里应该保留访问 ？
}

type CreateSessionRequestForArticles struct {
//列表页就不用 articleid 了
	StartTime  time.Time   `json:"start_time" binding:"required"`
	DeviceInfo *DeviceInfo `json:"device_info"`
}
// CreateSessionResponse 创建会话响应， 对应的 api
type CreateSessionResponse struct {
	SessionID uuid.UUID `json:"session_id"`
	UserID    uuid.UUID `json:"user_id"`
	ArticleID string       `json:"article_id"`
	StartTime time.Time `json:"start_time"`
}



