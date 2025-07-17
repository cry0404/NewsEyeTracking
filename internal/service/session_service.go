package service

import (
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/models"
	"context"
)

// SessionService 会话服务接口, 这里应该是针对阅读列表页和文章页的具体内容来做处理的服务层
type SessionService interface {
	// CreateSession 创建新的阅读会话
	CreateSessionForList(ctx context.Context, userID string, req *models.CreateSessionRequestForArticles) (*models.CreateSessionResponse, error)
	CreateSessionForFeed(ctx context.Context, userID string, req *models.CreateSessionRequestForArticle) (*models.CreateSessionResponse, error)
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

//使用的 req 还是一样的，默认将对列表页的 req 中的 articleid 设为0
func (s *sessionService) CreateSessionForList(ctx context.Context, userID string, req *models.CreateSessionRequestForArticles) (*models.CreateSessionResponse, error) {
 //为列表页单独处理的 session 服务
 //news + 当前年份 + 月日 + 0，列表默认为文章 0
	return nil, nil
}

func (s *sessionService) CreateSessionForFeed(ctx context.Context, userID string, req *models.CreateSessionRequestForArticle) (*models.CreateSessionResponse, error) {
	return nil, nil
}


// EndSession 实现结束会话逻辑
func (s *sessionService) EndSession(ctx context.Context, sessionID string, req *models.EndSessionRequest) error {
	//需要处理的是 endtime 和还残留的一些压缩数据
	return nil
}

