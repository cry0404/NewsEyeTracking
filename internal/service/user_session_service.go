package service

import (
	"NewsEyeTracking/internal/database"
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/models"
	"context"
	"encoding/json"
	"fmt"

	"time"

	"database/sql"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// UserSessionService 用户会话服务接口
type UserSessionService interface {
	// CreateOrGetUserSession 创建或获取用户的活跃会话
	CreateUserSession(ctx context.Context, userID uuid.UUID) (*models.CreateUserSessionResponse, error)
	// CheckSingleSessionLimit 检查单会话限制
	CheckSingleSessionLimit(ctx context.Context, userID uuid.UUID) error

	// Heartbeat 处理心跳请求
	//Heartbeat(ctx context.Context, req *models.HeartbeatRequest) (*models.HeartbeatResponse, error)

	// CleanupExpiredSessions 清理过期会话并写入数据库
	//CleanupExpiredSessions(ctx context.Context) error
}

// userSessionService 用户会话服务实现
type userSessionService struct {
	queries     *db.Queries
	redisClient *database.RedisClient
}

// NewUserSessionService 创建用户会话服务实例
func NewUserSessionService(queries *db.Queries, redisClient *database.RedisClient) UserSessionService {
	return &userSessionService{
		queries:     queries,
		redisClient: redisClient,
	}
}

const (
	// 新的简单结构：user_session:123e4567-e89b-12d3-a456-426614174000
	userSessionKeyPrefix = "user_session:"
	//activeUserKey        = "active_user" // 全局活跃用户键
	heartbeatTTL = 1 * time.Minute // 心跳TTL为1分钟
)

func (s *userSessionService) buildUserSessionKey(userID uuid.UUID) string {
	return fmt.Sprintf("%s%s", userSessionKeyPrefix, userID.String())
}

// CreateOrGetUserSession 创建或获取用户的活跃会话
func (s *userSessionService) CreateUserSession(ctx context.Context, userID uuid.UUID) (*models.CreateUserSessionResponse, error) {
	// 创建新会话, 初始的同时写入数据库, 然后最后更新就是了，没有 endtime 的就是意外退出事故
	now := time.Now()
	sessionID := uuid.New()

	newSessionData := models.RedisSessionData{
		ID:            sessionID,
		UserID:        userID,
		StartTime:     now,
		LastHeartbeat: now,
		IsActive:      true,
	}

	// 存储到Redis
	if err := s.saveSessionToRedis(ctx, newSessionData); err != nil {
		return nil, fmt.Errorf("保存会话到Redis失败")
	}

	err := s.queries.CreateUserSession(ctx, db.CreateUserSessionParams{
		ID:            newSessionData.ID,
		UserID:        newSessionData.UserID,
		StartTime:     sql.NullTime{Time: newSessionData.StartTime, Valid: true},
		LastHeartbeat: sql.NullTime{Time: newSessionData.LastHeartbeat, Valid: true},
		IsActive:      sql.NullBool{Bool: newSessionData.IsActive, Valid: true},
	})

	if err != nil {
		return nil, fmt.Errorf("将会话写入数据库失败: %w", err)
	}

	return &models.CreateUserSessionResponse{
		SessionID:     sessionID,
		UserID:        userID,
		StartTime:     now,
		LastHeartbeat: now,
		IsActive:      true,
	}, nil
} // 然后收到心跳包该更新 redis
//如何检测 redis 过期，在 ttl 之前将这次用户会话写入到数据库中去
// 但是如果又切回来该怎么算呢，是延续之前的会话，还是需要重新登录，或者说每次失去活性后就用刷新令牌刷新一下？
/*# 在redis.conf中添加：
notify-keyspace-events Ex */

// CheckSingleSessionLimit 在这里检查并创建应该也是可以的
func (s *userSessionService) CheckSingleSessionLimit(ctx context.Context, userID uuid.UUID) error {
	// 构建用户会话键
	userSessionKey := s.buildUserSessionKey(userID)

	// 尝试获取当前用户的会话
	sessionData, err := s.getSessionFromRedis(ctx, userSessionKey)
	if err != nil {
		if err.Error() == "会话不存在" {
			// 如果没有存在的会话，表示可以创建
			return nil
		}
		return fmt.Errorf("查询用户会话失败: %v", err)
	}

	// 如果存在活跃的会话且未过期，则返回错误，拒绝新的登录
	if sessionData.IsActive && time.Since(sessionData.LastHeartbeat) <= heartbeatTTL {
		return fmt.Errorf("用户已在其他位置登录")
	}

	// 否则，可以创建新的会话
	return nil
}

/*
// getSessionByUserID 根据用户ID获取会话数据
func (s *userSessionService) getSessionByUserID(ctx context.Context, userID uuid.UUID) (*models.RedisSessionData, error) {
	userSessionKey := s.buildUserSessionKey(userID)
	return s.getSessionFromRedis(ctx, userSessionKey)
}*/

// getSessionFromRedis 从Redis获取会话数据
func (s *userSessionService) getSessionFromRedis(ctx context.Context, key string) (*models.RedisSessionData, error) {
	data, err := s.redisClient.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("会话不存在")
		}
		return nil, err
	}

	var sessionData models.RedisSessionData
	if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
		return nil, fmt.Errorf("解析会话数据失败: %v", err)
	}

	return &sessionData, nil
}

// saveSessionToRedis 保存会话数据到Redis
func (s *userSessionService) saveSessionToRedis(ctx context.Context, sessionData models.RedisSessionData) error {
	key := s.buildUserSessionKey(sessionData.UserID)

	data, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("序列化会话数据失败: %v", err)
	}

	return s.redisClient.Set(ctx, key, string(data), heartbeatTTL)
}

/*
// createUserSessionInDB 在数据库中创建会话记录（用于数据关联）,这个部分最后做
func (s *userSessionService) createUserSessionInDB(ctx context.Context, sessionData RedisSessionData) error {
	// 这里需要等待数据库迁移完成后实现
	// 暂时返回nil
	return nil
}*/
