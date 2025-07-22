package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EyeEvent 眼动事件数据
type EyeEvent struct {
	ID string `json:"id"` // 元素ID（分词或组件的唯一标识）
	X  float32 `json:"x"`  // X坐标
	Y  float32 `json:"y"`  // Y坐标
}

// ClickEvent 点击事件数据
type ClickEvent struct {
	Timestamp time.Time  `json:"timestamp"`           // Unix时间戳（毫秒）
	ID        string    `json:"id"`                  // 点击目标的元素ID
	X         float32    `json:"x"`                   // 点击的X坐标
	Y         float32    `json:"y"`                   // 点击的Y坐标
}

// ScrollEvent 滚动事件数据
type ScrollEvent struct {
	Timestamp time.Time `json:"timestamp"` // Unix时间戳（毫秒）
	DeltaY    float32  `json:"delta_y"`   // 垂直滚动距离（正数向下，负数向上）
}

// TrackingData 追踪数据容器
type TrackingData struct {
	EyeEvents    []EyeEvent    `json:"eye_event,omitempty"`
	ClickEvents  []ClickEvent  `json:"click_event,omitempty"`
	ScrollEvents []ScrollEvent `json:"scroll_event,omitempty"`
}

// SessionDataRequest 会话数据请求（支持心跳包和数据传输的混合格式）
// session_id 从 URL 参数获取，不需要在请求体中包含
type SessionDataRequest struct {
	SessionID *uuid.UUID    `json:"-"` // 内部使用，从URL参数设置，不序列化到JSON
	Data      *TrackingData `json:"data,omitempty"`       // 追踪数据（数据传输时包含）
	Ping      *bool         `json:"ping,omitempty"`       // 心跳包标识（心跳时为 true）
	Timestamp time.Time     `json:"timestamp"`            // 请求时间戳
}

// EndSessionRequest 结束会话请求
// session_id 从 URL 参数获取，不需要在请求体中包含
type EndSessionRequest struct {
	EndTime time.Time     `json:"end_time" binding:"required"`
	Data    *TrackingData `json:"data,omitempty"` // 最后一批追踪数据
}

// UserTrackingRecord 用户追踪记录（用于缓存）
type UserTrackingRecord struct {
	SessionID uuid.UUID    `json:"session_id"`
	StartTime time.Time    `json:"start_time"`
	Data      TrackingData `json:"data"`
}

// UserTrackingBatchRecord 用户追踪批量记录（用于文件存储）
type UserTrackingBatchRecord struct {
	FlushTime time.Time            `json:"flush_time"`
	Records   []UserTrackingRecord `json:"records"`
}

// IsHeartbeat 检查是否为心跳包
func (r *SessionDataRequest) IsHeartbeat() bool {
	return r.Ping != nil && *r.Ping
}

// HasTrackingData 检查是否包含追踪数据
// SessionID 从 URL 参数设置，只需要检查 Data 字段
func (r *SessionDataRequest) HasTrackingData() bool {
	return r.Data != nil && !r.Data.IsEmpty()
}

// IsEmpty 检查追踪数据是否为空
func (td *TrackingData) IsEmpty() bool {
	return len(td.EyeEvents) == 0 && len(td.ClickEvents) == 0 && len(td.ScrollEvents) == 0
}

// Validate 验证请求数据的有效性
func (r *SessionDataRequest) Validate() error {
	// 心跳包验证
	if r.IsHeartbeat() {
		return nil // 心跳包不需要其他数据验证
	}

	// 数据传输验证（session_id 从 URL 参数获取，不需要在请求体中验证）
	if r.HasTrackingData() {
		return nil // 有追踪数据就是有效的
	}

	// 如果既不是心跳包也没有追踪数据，则为无效请求
	return fmt.Errorf("request must be either a heartbeat or contain tracking data")
}

// GetDataSummary 获取数据摘要（用于日志记录）
func (r *SessionDataRequest) GetDataSummary() string {
	if r.IsHeartbeat() {
		return "heartbeat"
	}

	if r.Data == nil {
		return "no_data"
	}

	eyeCount := len(r.Data.EyeEvents)
	clickCount := len(r.Data.ClickEvents)
	scrollCount := len(r.Data.ScrollEvents)

	return fmt.Sprintf("eye:%d,click:%d,scroll:%d", eyeCount, clickCount, scrollCount)
}

// GetTotalEvents 获取事件总数
func (td *TrackingData) GetTotalEvents() int {
	return len(td.EyeEvents) + len(td.ClickEvents) + len(td.ScrollEvents)
}

// AddEyeEvent 添加眼动事件
func (td *TrackingData) AddEyeEvent(id string, x, y float32) {
	td.EyeEvents = append(td.EyeEvents, EyeEvent{
		ID: id,
		X:  x,
		Y:  y,
	})
}

// AddClickEvent 添加点击事件
func (td *TrackingData) AddClickEvent(timestamp time.Time, id string, x, y float32) {
	td.ClickEvents = append(td.ClickEvents, ClickEvent{
		Timestamp: timestamp,
		ID:        id,
		X:         x,
		Y:         y,
	})
}

// AddScrollEvent 添加滚动事件
func (td *TrackingData) AddScrollEvent(timestamp time.Time, deltaY float32) {
	td.ScrollEvents = append(td.ScrollEvents, ScrollEvent{
		Timestamp: timestamp,
		DeltaY:    deltaY,
	})
}

// SessionDataResponse 会话数据响应（支持心跳包和数据传输的混合响应）
type SessionDataResponse struct {
	SessionID *uuid.UUID `json:"session_id,omitempty"` // 数据传输成功时返回会话ID
	Success   *bool      `json:"success,omitempty"`    // 数据传输成功标识
	Pong      *bool      `json:"pong,omitempty"`       // 心跳包响应标识
	Timestamp time.Time  `json:"timestamp"`            // 响应时间戳
}

// NewDataResponse 创建数据传输成功响应
func NewDataResponse(sessionID uuid.UUID) *SessionDataResponse {
	success := true
	return &SessionDataResponse{
		SessionID: &sessionID,
		Success:   &success,
		Timestamp: time.Now(),
	}
}

// NewHeartbeatResponse 创建心跳包响应
func NewHeartbeatResponse() *SessionDataResponse {
	pong := true
	return &SessionDataResponse{
		Pong:      &pong,
		Timestamp: time.Now(),
	}
}

// IsDataResponse 检查是否为数据传输响应
func (r *SessionDataResponse) IsDataResponse() bool {
	return r.Success != nil
}

// IsHeartbeatResponse 检查是否为心跳包响应
func (r *SessionDataResponse) IsHeartbeatResponse() bool {
	return r.Pong != nil && *r.Pong
}
