package service

import (
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/models"
	"context"
)

// SessionService 会话服务接口, 这里应该是针对阅读列表页和文章页的具体内容来做处理的服务层
type SessionService interface {
	// CreateSession 创建新的阅读会话
	CreateSession(ctx context.Context, userID string, req *models.CreateSessionRequest) (*models.CreateSessionResponse, error)
	// EndSession 结束阅读会话
	EndSession(ctx context.Context, sessionID string, req *models.EndSessionRequest) error
	
}

// sessionService 会话服务实现
type sessionService struct {
	queries *db.Queries
}

// NewSessionService 创建会话服务实例
func NewSessionService(queries *db.Queries) SessionService {
	return &sessionService{queries: queries}
}

// CreateSession 实现创建会话逻辑
func (s *sessionService) CreateSession(ctx context.Context, userID string, req *models.CreateSessionRequest) (*models.CreateSessionResponse, error) {
	// TODO: 实现创建会话逻辑
	// 1. 验证文章ID是否存在
	// 2. 创建会话记录
	// 3. 返回会话信息
	return nil, nil
}

// EndSession 实现结束会话逻辑
func (s *sessionService) EndSession(ctx context.Context, sessionID string, req *models.EndSessionRequest) error {
	// TODO: 实现结束会话逻辑
	// 1. 更新会话结束时间
	// 2. 如果有压缩数据，上传到OSS
	// 3. 更新会话统计信息
	return nil
}

