package utils

import (
	"context"
	"time"
)

// 定义不同操作的超时时间
const (
	// 数据库查询超时时间
	DatabaseQueryTimeout = 5 * time.Second
	
	// 复杂查询超时时间（如分析类查询）
	ComplexQueryTimeout = 10 * time.Second
	
	// 写操作超时时间
	WriteOperationTimeout = 10 * time.Second
	
	// 认证操作超时时间
	AuthOperationTimeout = 2 * time.Second
	
	// 外部API调用超时时间
	ExternalAPITimeout = 15 * time.Second
	
	// 文件操作超时时间
	FileOperationTimeout = 30 * time.Second
)

// WithDatabaseTimeout 为数据库查询创建带超时的 context
func WithDatabaseTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, DatabaseQueryTimeout)
}

// WithComplexQueryTimeout 为复杂查询创建带超时的 context
func WithComplexQueryTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, ComplexQueryTimeout)
}

// WithReadTimeout 为读操作创建带超时的 context
func WithReadTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, DatabaseQueryTimeout)
}

// WithWriteTimeout 为写操作创建带超时的 context
func WithWriteTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, WriteOperationTimeout)
}

// WithAuthTimeout 为认证操作创建带超时的 context
func WithAuthTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, AuthOperationTimeout)
}

// WithExternalAPITimeout 为外部API调用创建带超时的 context
func WithExternalAPITimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, ExternalAPITimeout)
}

// WithFileOperationTimeout 为文件操作创建带超时的 context
func WithFileOperationTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, FileOperationTimeout)
}

// WithCustomTimeout 创建自定义超时的 context
func WithCustomTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}
