package service

import (
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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
	// ForceEndSession 强制结束阅读会话（用于清理服务）
	ForceEndSession(ctx context.Context, sessionID uuid.UUID) error
	// GetSessionByID 根据ID获取会话信息（用于验证会话状态）
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*db.ReadingSession, error)
	// GetUserActiveSessions 获取用户所有活跃的阅读会话
	GetUserActiveSessions(ctx context.Context, userID uuid.UUID) ([]db.ReadingSession, error)
	// CheckActiveReadingSession 检查用户是否已有活跃的阅读会话（单会话限制）
	//CheckActiveReadingSession(ctx context.Context, userID string) error
}

// 现在需要思考的是打开一个新的会话就相当于关闭一个会话是吗
// sessionService 会话服务实现
type sessionService struct {
	queries     *db.Queries
	//redisClient *database.RedisClient
}

// NewSessionService 创建会话服务实例
func NewSessionService(queries *db.Queries) SessionService {
	return &sessionService{
		queries:     queries,
		//redisClient: redisClient, //默认为 0 ，考虑 gjc 那边接入推荐算法的缓存

	}
}

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
		UserID:    userUUID,
		ArticleID: listArticleID, // 使用字符串格式的列表页ID
		StartTime: sql.NullTime{Time: req.StartTime, Valid: true},
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

	//if err := s.saveSessionToRedis(ctx, ReadingsessionData models.RedisReadingSessionData)
	//创建成功后在这里构建 redis 缓存, 调用函数吧，都是阅读会话

	// 构建响应，返回格式化的列表页ID
	response := &models.CreateSessionResponse{
		SessionID: session.ID,
		UserID:    session.UserID,
		ArticleID: listArticleID, // 返回格式化的列表页ID: news20250000
		StartTime: session.StartTime.Time,
	}/*
	RedisReadingSessionData := models.RedisReadingSessionData{
		ReadingSessionID: session.ID,
		UserID:           session.UserID,
		ArticleID:        listArticleID,
		StartTime:        session.StartTime.Time,
		EndTime:          nil,
	}*/
	//s.saveSessionToRedis(ctx, RedisReadingSessionData)
	return response, nil
}

func (s *sessionService) CreateSessionForFeed(ctx context.Context, userID string, req *models.CreateSessionRequestForArticle) (*models.CreateSessionResponse, error) {

	/*if err := s.CheckActiveReadingSession(ctx, userID); err != nil {
		return nil, err
	}*/

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
		UserID:    userUUID,
		ArticleID: req.ArticleID,
		StartTime: sql.NullTime{Time: req.StartTime, Valid: true},
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
	}/*
	RedisReadingSessionData := models.RedisReadingSessionData{
		ReadingSessionID: session.ID,
		UserID:           session.UserID,
		ArticleID:        req.ArticleID,
		StartTime:        session.StartTime.Time,
		EndTime:          nil,
	}
	s.saveSessionToRedis(ctx, RedisReadingSessionData)*/
	return response, nil
}

// EndReadingSession 结束阅读会话，先删除 Redis 缓存再更新数据库
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
/*
	// 第一步：先删除 Redis 中的阅读会话缓存
	if err := s.DeleteSessionInRedis(ctx, existingSession.UserID); err != nil {
		// Redis 删除失败记录日志，但不阻断流程
		log.Printf("删除 Redis 阅读会话缓存失败: %v", err)
	}*/

	sessionInfo := db.UpdateSessionEndTimeParams{
		ID:      sessionUUID,
		EndTime: sql.NullTime{Time: req.EndTime, Valid: true},
	}

	err = s.queries.UpdateSessionEndTime(ctx, sessionInfo)
	if err != nil {
		return fmt.Errorf("更新阅读会话结束时间失败: %w", err)
	}

	log.Printf("成功结束阅读会话，会话ID: %s，用户ID: %s", sessionID, existingSession.UserID.String())
	return nil
}

// GetSessionByID 根据ID获取会话信息（用于验证会话状态）
func (s *sessionService) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*db.ReadingSession, error) {
	session, err := s.queries.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// ForceEndSession 强制结束阅读会话（用于清理服务）
func (s *sessionService) ForceEndSession(ctx context.Context, sessionID uuid.UUID) error {
	// 验证会话是否存在
	existingSession, err := s.queries.GetSessionByID(ctx, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("阅读会话不存在")
		}
		return fmt.Errorf("查询阅读会话失败: %w", err)
	}

	// 如果会话已经结束，直接返回
	if existingSession.EndTime.Valid {
		log.Printf("阅读会话 %s 已经结束，无需重复处理", sessionID.String())
		return nil
	}

	// 强制结束会话，使用当前时间作为结束时间
	sessionInfo := db.UpdateSessionEndTimeParams{
		ID:      sessionID,
		EndTime: sql.NullTime{Time: time.Now(), Valid: true},
	}

	err = s.queries.UpdateSessionEndTime(ctx, sessionInfo)
	if err != nil {
		return fmt.Errorf("强制结束阅读会话失败: %w", err)
	}

	log.Printf("成功强制结束阅读会话，会话 ID: %s，用户 ID: %s", 
		sessionID.String(), existingSession.UserID.String())
	return nil
}

// GetUserActiveSessions 获取用户所有活跃的阅读会话
func (s *sessionService) GetUserActiveSessions(ctx context.Context, userID uuid.UUID) ([]db.ReadingSession, error) {
	activeSessions, err := s.queries.GetUserActiveSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户活跃阅读会话失败: %w", err)
	}
	return activeSessions, nil
}

/*
const (
	//主要是有没有必要，似乎只有最后才会更新阅读会话，也与频繁读取关系不大？
	readingSessionKeyPrefix = "reading_session:"
	DefaultTTL              = 5 * time.Minute //先按照一篇正常的新闻阅读时间来
)

func (s *sessionService) buildReadingSessionKey(userID uuid.UUID) string { // 加上前缀后就是唯一的
	return fmt.Sprintf("%s%s", readingSessionKeyPrefix, userID.String())
} //查询存储删除

func (s *sessionService) getSessionFromRedis(ctx context.Context, key string) (*models.RedisReadingSessionData, error) {
	data, err := s.redisClient.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("会话不存在")
		}
		return nil, err
	}
	var sessionData models.RedisReadingSessionData
	if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
		return nil, fmt.Errorf("解析会话数据失败")
	}
	return &sessionData, nil
}

func (s *sessionService) saveSessionToRedis(ctx context.Context, ReadingsessionData models.RedisReadingSessionData) error {
	key := s.buildReadingSessionKey(ReadingsessionData.ReadingSessionID)

	data, err := json.Marshal(ReadingsessionData)
	if err != nil {
		return fmt.Errorf("redis 中序列化会话数据失败")
	}


	return s.redisClient.Set(ctx, key, string(data), DefaultTTL)
}

// DeleteSessionInRedis 从 Redis 中删除阅读会话缓存
func (s *sessionService) DeleteSessionInRedis(ctx context.Context, userID uuid.UUID) error {
	key := s.buildReadingSessionKey(userID)
	
	// 检查键是否存在
	exists, err := s.redisClient.Exists(ctx, key)
	if err != nil {
		return fmt.Errorf("检查 Redis 键存在性失败: %w", err)
	}
	
	// 如果键不存在，不需要删除
	if !exists {
		log.Printf("Redis 中不存在阅读会话缓存，键: %s", key)
		return nil
	}
	
	// 删除 Redis 中的缓存
	if err := s.redisClient.Delete(ctx, key); err != nil {
		return fmt.Errorf("从 Redis 删除阅读会话缓存失败: %w", err)
	}
	
	log.Printf("成功从 Redis 删除阅读会话缓存，键: %s", key)
	return nil
}

// CheckActiveReadingSession 检查用户是否已有活跃的阅读会话（单会话限制）， 可以改为从 redis 中先查询, 检查时从 redis 中查看，创建时在 redis 中创建缓存
func (s *sessionService) CheckActiveReadingSession(ctx context.Context, userID string) error {
	// 解析用户ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("用户ID格式不正确: %w", err)
	}
	//先从缓存查， 按道理来说这里应该不读取数据库
	redisKey := s.buildReadingSessionKey(userUUID)
	redisSessions, err := s.getSessionFromRedis(ctx, redisKey) // 键值不为空就应该删除啦
	if err != nil {
		log.Printf("从 redis 中获取缓存失败: %v", err)
	}

	if redisSessions != nil {
		return fmt.Errorf("用户已有活跃会话，请先结束当前阅读会话")
	}
	// 查询用户的活跃阅读会话（end_time 为 NULL 的会话）
	activeSessions, err := s.queries.GetUserActiveSessions(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("查询活跃会话失败: %w", err)
	}

	// 如果存在活跃会话，则禁止创建新会话
	if len(activeSessions) > 0 {
		return fmt.Errorf("用户已有活跃的阅读会话，请先结束当前阅读会话")
	}

	return nil
}
*/
// 使用的 req 还是一样的，默认将对列表页的 req 中的 articleid 设为0
