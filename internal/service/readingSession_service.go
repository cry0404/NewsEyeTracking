package service

import (
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)
// 感觉也可以为 阅读会话添加对应的键
// SessionService 会话服务接口, 这里应该是针对阅读列表页和文章页的具体内容来做处理的服务层
type SessionService interface {
	// CreateSession 创建新的阅读会话
	CreateSessionForList(ctx context.Context, userID string, req *models.CreateSessionRequestForArticles) (*models.CreateSessionResponse, error)
	CreateSessionForFeed(ctx context.Context, userID string, req *models.CreateSessionRequestForArticle) (*models.CreateSessionResponse, error)
	// EndSession 结束阅读会话
	EndReadingSession(ctx context.Context, sessionID string, req *models.EndSessionRequest) error
	// CheckActiveReadingSession 检查用户是否已有活跃的阅读会话（单会话限制）
	CheckActiveReadingSession(ctx context.Context, userID string) error
	
}
// 现在需要思考的是打开一个新的会话就相当于关闭一个会话是吗
// sessionService 会话服务实现
type sessionService struct {
	queries *db.Queries
}

// NewSessionService 创建会话服务实例
func NewSessionService(queries *db.Queries) SessionService {
	return &sessionService{queries: queries}
}
/*
const (
	主要是有没有必要，似乎只有最后才会更新阅读会话，也与频繁读取关系不大？
	readingSessionKeyPrefix = "reading_session"
)*/
//使用的 req 还是一样的，默认将对列表页的 req 中的 articleid 设为0
func (s *sessionService) CreateSessionForList(ctx context.Context, userID string, req *models.CreateSessionRequestForArticles) (*models.CreateSessionResponse, error) {
	// 检查用户是否已有活跃的阅读会话（单会话限制）
	if err := s.CheckActiveReadingSession(ctx, userID); err != nil {
		return nil, err
	}
	
	// 为列表页单独处理的 session 服务
	// news + 当前年份 + 月日 + 000，列表默认为文章 000
	
	// 解析用户ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("用户ID格式不正确: %w", err)
	}
	
	// 生成列表页专用的文章ID格式: news20250718000
	currentTime := time.Now()
	listArticleID := fmt.Sprintf("news%s000", currentTime.Format("20060102"))
	
	// 序列化设备信息
	var deviceInfoJSON []byte
	if req.DeviceInfo != nil {
		deviceInfoJSON, err = json.Marshal(req.DeviceInfo)
		if err != nil {
			return nil, fmt.Errorf("序列化设备信息失败: %w", err)
		}
	}
	
	// 创建会话参数
	sessionParams := db.CreateSessionParams{
		UserID:     userUUID,
		ArticleID:  listArticleID, // 使用字符串格式的列表页ID
		StartTime:  sql.NullTime{Time: req.StartTime, Valid: true},
		DeviceInfo: pqtype.NullRawMessage{
			RawMessage: deviceInfoJSON,
			Valid:      req.DeviceInfo != nil,
		},
	}
	
	// 创建会话
	session, err := s.queries.CreateSession(ctx, sessionParams)
	if err != nil {
		return nil, fmt.Errorf("创建列表页会话失败: %w", err)
	}
	
	// 构建响应，返回格式化的列表页ID
	response := &models.CreateSessionResponse{
		SessionID: session.ID,
		UserID:    session.UserID,
		ArticleID: listArticleID, // 返回格式化的列表页ID: news20250000
		StartTime: session.StartTime.Time,
	}
	
	return response, nil
}

func (s *sessionService) CreateSessionForFeed(ctx context.Context, userID string, req *models.CreateSessionRequestForArticle) (*models.CreateSessionResponse, error) {
	// 1. 检查用户是否已有活跃的阅读会话（单会话限制）
	if err := s.CheckActiveReadingSession(ctx, userID); err != nil {
		return nil, err
	}
	
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("用户 id 格式不合法")
	}


	var deviceInfoJSON []byte
	if req.DeviceInfo != nil {
		deviceInfoJSON, err = json.Marshal(req.DeviceInfo)
		if err != nil {
			return nil, fmt.Errorf("序列化设备信息失败")
		}
	}


	sessionParams := db.CreateSessionParams{
		UserID:     userUUID,
		ArticleID:  req.ArticleID,
		StartTime:  sql.NullTime{Time: req.StartTime, Valid: true},
		DeviceInfo: pqtype.NullRawMessage{
			RawMessage: deviceInfoJSON,
			Valid:      req.DeviceInfo != nil,
		},
	}


	session, err := s.queries.CreateSession(ctx, sessionParams)
	if err != nil {
		return nil, fmt.Errorf("创建会话失败")
	}

	response := &models.CreateSessionResponse{
		SessionID: session.ID,
		UserID:    session.UserID,
		ArticleID: req.ArticleID,
		StartTime: session.StartTime.Time,
	}

	return response, nil
}


// EndReadingSession 结束阅读会话，包含完整的验证逻辑
func (s *sessionService) EndReadingSession(ctx context.Context, sessionID string, req *models.EndSessionRequest) error {
	//  验证并解析会话ID
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("会话ID格式不正确: %w", err)
	}

	//  验证会话是否存在
	existingSession, err := s.queries.GetSessionByID(ctx, sessionUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("阅读会话不存在")
		}
		return fmt.Errorf("查询阅读会话失败: %w", err)
	}

	//  验证会话是否已经结束
	if existingSession.EndTime.Valid {
		return fmt.Errorf("阅读会话已经结束")
	}

	//  更新会话结束时间
	sessionInfo := db.UpdateSessionEndTimeParams{
		ID:      sessionUUID,
		EndTime: sql.NullTime{Time: req.EndTime, Valid: true},
	}

	err = s.queries.UpdateSessionEndTime(ctx, sessionInfo)
	if err != nil {
		return fmt.Errorf("更新阅读会话结束时间失败: %w", err)
	}

	
	// - 将追踪数据持久化到文件或其他存储
	// - 发送会话结束事件

	return nil
}

// CheckActiveReadingSession 检查用户是否已有活跃的阅读会话（单会话限制）， 可以改为从 redis 中先查询
func (s *sessionService) CheckActiveReadingSession(ctx context.Context, userID string) error {
	// 解析用户ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("用户ID格式不正确: %w", err)
	}

	// 查询用户的活跃阅读会话（end_time 为 NULL 的会话）
	activeSessions, err := s.queries.GetUserActiveSessions(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("查询活跃会话失败: %w", err)
	}

	// 如果存在活跃会话，则禁止创建新会话
	if len(activeSessions) > 0 {
		return fmt.Errorf("用户已有活跃的阅读会话，请先结束当前会话")
	}

	return nil
}

