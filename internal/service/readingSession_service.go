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

// SessionService 会话服务接口, 这里应该是针对阅读列表页和文章页的具体内容来做处理的服务层
type SessionService interface {
	// CreateSession 创建新的阅读会话
	CreateSessionForList(ctx context.Context, userID string, req *models.CreateSessionRequestForArticles) (*models.CreateSessionResponse, error)
	CreateSessionForFeed(ctx context.Context, userID string, req *models.CreateSessionRequestForArticle) (*models.CreateSessionResponse, error)
	// EndSession 结束阅读会话
	EndReadingSession(ctx context.Context, sessionID string, req *models.EndSessionRequest) error
	
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

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("用户 id 格式不合法“")
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


// EndSession 实现结束会话逻辑, 也需要一个对应的路径
func (s *sessionService) EndReadingSession(ctx context.Context, sessionID string, req *models.EndSessionRequest) error {
	sessionid, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("解析失败，请检查 session_id 格式是否正确")
	}
	sessionInfo := db.UpdateSessionEndTimeParams{
		ID: 	sessionid,
		EndTime: sql.NullTime{Time: req.EndTime, Valid: true},
	}
	err = s.queries.UpdateSessionEndTime(ctx, sessionInfo)
	if err != nil {
		return fmt.Errorf("更新会话结束时间失败")
	}

	//这里的存储就要调用解析存储的逻辑函数了
	return nil
}

