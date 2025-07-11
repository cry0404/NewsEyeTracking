package models

import (
	"time"
)

// EventType 事件类型枚举
type EventType string

const (
	EventTypeEye    EventType = "eye"    // 眼动事件
	EventTypeScroll EventType = "scroll" // 滚动事件
	EventTypeClick  EventType = "click"  // 点击事件
)

// EyeEvent 眼动事件
type EyeEvent struct {
	ClassID string `json:"class_id"` // 文本分类ID
	X       int    `json:"x"`        // X坐标
	Y       int    `json:"y"`        // Y坐标
}

// ScrollEvent 滚动事件
type ScrollEvent struct {
	TimestampInterval int64 `json:"timestamp_interval"` // 时间间隔（毫秒）
	ScrollDistance    int   `json:"scroll_distance"`    // 滚动距离（正数向下，负数向上）
}

// ClickEvent 点击事件
type ClickEvent struct {
	TimestampInterval int64 `json:"timestamp_interval"` // 时间间隔（毫秒）
	X                 int   `json:"x"`                  // X坐标
	Y                 int   `json:"y"`                  // Y坐标
}

// ParsedEvent 解析后的事件（包含时间戳）
type ParsedEvent struct {
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"` // EyeEvent, ScrollEvent, 或 ClickEvent
}

// EventBatch 事件批次（同一时间点的多个事件）
type EventBatch struct {
	Timestamp time.Time     `json:"timestamp"`
	Events    []ParsedEvent `json:"events"`
}

// CompressedDataStats 压缩数据统计
type CompressedDataStats struct {
	EyeEvents    int `json:"eye_events"`
	ScrollEvents int `json:"scroll_events"`
	ClickEvents  int `json:"click_events"`
	TotalEvents  int `json:"total_events"`
	DataSize     int `json:"data_size"` // 原始字符串大小
}

// DataProcessor 数据处理器接口
type DataProcessor interface {
	// ParseCompressedData 解析压缩的数据字符串
	ParseCompressedData(compressedData string) ([]EventBatch, error)

	// CompressEvents 将事件压缩为字符串
	CompressEvents(batches []EventBatch) (string, error)

	// GetStats 获取数据统计信息
	GetStats(compressedData string) (*CompressedDataStats, error)
}

// OSSUploadInfo OSS上传信息
type OSSUploadInfo struct {
	Bucket   string `json:"bucket"`
	Key      string `json:"key"`      // 文件路径
	URL      string `json:"url"`      // 访问URL
	Size     int64  `json:"size"`     // 文件大小
	Checksum string `json:"checksum"` // 文件校验和
}
