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

type LoginResponse struct {
	SessionID uuid.UUID `json:"session_id"`
	StartTime time.Time `json:"start_time"`
	Token     string    `json:"token"`
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

// SessionStatusResponse 会话状态响应
type SessionStatusResponse struct {
	SessionID        uuid.UUID `json:"session_id"`
	IsActive         bool      `json:"is_active"`
	LastHeartbeat    time.Time `json:"last_heartbeat"`
	IsExpired        bool      `json:"is_expired"`
	CanCreateReading bool      `json:"can_create_reading"` // 是否可以创建阅读会话
}

// RedisSessionData Redis中存储的会话数据，使用UserSession的结构
type RedisUserSessionData UserSession
