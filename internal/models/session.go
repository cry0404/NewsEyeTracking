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

	// OSS存储相关
	OSSFilePath *string `json:"oss_file_path" db:"oss_file_path"`
	DataSize    int64   `json:"data_size" db:"data_size"`
	EventCount  int     `json:"event_count" db:"event_count"`

	// 会话统计信息
	TotalEyeEvents    int    `json:"total_eye_events" db:"total_eye_events"`
	TotalClickEvents  int    `json:"total_click_events" db:"total_click_events"`
	TotalScrollEvents int    `json:"total_scroll_events" db:"total_scroll_events"`
	SessionDurationMs *int64 `json:"session_duration_ms" db:"session_duration_ms"`
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	ArticleID  int         `json:"article_id" binding:"required"`
	StartTime  time.Time   `json:"start_time" binding:"required"`
	DeviceInfo *DeviceInfo `json:"device_info"`
}

// CreateSessionResponse 创建会话响应， 对应的 api
type CreateSessionResponse struct {
	SessionID uuid.UUID `json:"session_id"`
	UserID    uuid.UUID `json:"user_id"`
	ArticleID int       `json:"article_id"`
	StartTime time.Time `json:"start_time"`
}

// EndSessionRequest 结束会话请求
type EndSessionRequest struct {
	EndTime        time.Time `json:"end_time" binding:"required"`
	CompressedData *string   `json:"compressed_data"` // 最后一批压缩数据
}

// UploadCompressedDataRequest 上报压缩数据请求
type UploadCompressedDataRequest struct {
	CompressedData string `json:"compressed_data" binding:"required"`
}

// UploadCompressedDataResponse 上报压缩数据响应
type UploadCompressedDataResponse struct {
	SessionID    uuid.UUID `json:"session_id"`
	UploadedSize int64     `json:"uploaded_size"`
	EventCount   int       `json:"event_count"`
}

// UploadDataRequest 上传压缩数据请求
type UploadDataRequest struct {
	CompressedData string `json:"compressed_data" binding:"required"`
}

// UploadDataResponse 上传压缩数据响应
type UploadDataResponse struct {
	SessionID    string `json:"session_id"`
	UploadedSize int64  `json:"uploaded_size"`
	EventCount   int    `json:"event_count"`
}
