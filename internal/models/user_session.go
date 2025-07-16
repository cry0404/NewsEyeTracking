package models

import (
	"time"

	"github.com/google/uuid"
)

// UserSession 用户登录会话模型
type UserSession struct {
	ID            uuid.UUID `json:"id" db:"id"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	StartTime     time.Time `json:"start_time" db:"start_time"`
	LastHeartbeat time.Time `json:"last_heartbeat" db:"last_heartbeat"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	CreatedDate   time.Time `json:"created_date" db:"created_date"`
}

// CreateUserSessionRequest 创建用户会话请求
type CreateUserSessionRequest struct {
	UserID    uuid.UUID `json:"user_id" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
}

// CreateUserSessionResponse 创建用户会话响应
type CreateUserSessionResponse struct {
	SessionID     uuid.UUID `json:"session_id"`
	UserID        uuid.UUID `json:"user_id"`
	StartTime     time.Time `json:"start_time"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	IsActive      bool      `json:"is_active"`
}

// HeartbeatRequest 心跳请求
type HeartbeatRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
	Timestamp time.Time `json:"timestamp" binding:"required"`
}

// HeartbeatResponse 心跳响应
type HeartbeatResponse struct {
	SessionID     uuid.UUID `json:"session_id"`
	Status        string    `json:"status"` // "ok", "expired", "invalid"
	LastHeartbeat time.Time `json:"last_heartbeat"`
	Message       string    `json:"message,omitempty"`
}

// DataUploadRequest 数据上传请求（包含心跳检测）
type DataUploadRequest struct {
	SessionID      uuid.UUID `json:"session_id" binding:"required"`
	DataType       string    `json:"data_type" binding:"required"` // "heartbeat", "eyetracking", "mixed"
	CompressedData string    `json:"compressed_data" binding:"required"`
	Timestamp      time.Time `json:"timestamp" binding:"required"`
}

// DataUploadResponse 数据上传响应
type DataUploadResponse struct {
	SessionID       uuid.UUID `json:"session_id"`
	ProcessedType   string    `json:"processed_type"` // "heartbeat_only", "eyetracking_only", "mixed"
	HeartbeatStatus string    `json:"heartbeat_status,omitempty"`
	DataSize        int64     `json:"data_size,omitempty"`
	EventCount      int       `json:"event_count,omitempty"`
	Message         string    `json:"message,omitempty"`
}

// SessionStatusResponse 会话状态响应
type SessionStatusResponse struct {
	SessionID       uuid.UUID `json:"session_id"`
	IsActive        bool      `json:"is_active"`
	LastHeartbeat   time.Time `json:"last_heartbeat"`
	IsExpired       bool      `json:"is_expired"`
	CanCreateReading bool     `json:"can_create_reading"` // 是否可以创建阅读会话
}
