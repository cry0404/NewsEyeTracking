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
	// CreateUserSession 创建新的用户会话
	CreateUserSession(ctx context.Context, userID uuid.UUID) (*models.CreateUserSessionResponse, error)
	// CheckSingleSessionLimit 检查用户的单会话限制
	CheckSingleSessionLimit(ctx context.Context, userID uuid.UUID) error
	// Heartbeat 处理心跳请求，同时更新 Redis 和数据库
	Heartbeat(ctx context.Context, req *models.HeartbeatRequest) (*models.HeartbeatResponse, error)
	// CleanupExpiredSessions 清理过期会话
	CleanupExpiredSessions(ctx context.Context) error
	// GetSessionStatus 获取会话状态
	GetSessionStatus(ctx context.Context, sessionID uuid.UUID) (*models.SessionStatusResponse, error)
	// EndSession 手动结束会话
	EndUserSession(ctx context.Context, sessionID uuid.UUID) error
	// GetActiveUserSessionByUserID 根据用户ID获取活跃会话
	GetActiveUserSessionByUserID(ctx context.Context, userID uuid.UUID) (*db.GetActiveUserSessionByUserIDRow, error)
}

// userSessionService 用户会话服务实现
// redis 针对用户会话主要应该来处理
type userSessionService struct {
	queries     *db.Queries
	redisClient *database.RedisClient
}

// NewUserSessionService 创建用户会话服务实例
func NewUserSessionService(queries *db.Queries, redisClient *database.RedisClient) UserSessionService {
	return &userSessionService{
		queries:     queries,
		redisClient: redisClient, //默认数据库为 0
	}
}

const (
	// user_id + session_id 的键怎么样，但是确定不了
	// 新的简单结构：user_session:123e4567-e89b-12d3-a456-426614174000
	//定义 reading_session 单会话呢
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
// 收到心跳包更新两个会话的时间限制吗？
/*# 在redis.conf中添加：
notify-keyspace-events Ex */

// CheckSingleSessionLimit 检查用户的单会话限制（只使用Redis）
func (s *userSessionService) CheckSingleSessionLimit(ctx context.Context, userID uuid.UUID) error {
	// 构建用户会话键
	userSessionKey := s.buildUserSessionKey(userID)

	// 尝试从 Redis 获取当前用户的会话
	sessionData, err := s.getSessionFromRedis(ctx, userSessionKey)
	if err != nil {
		if err.Error() == "会话不存在" {
			// Redis 中没有活跃会话，可以创建新会话
			return nil
		}
		return fmt.Errorf("查询 Redis 失败")
	}

	// 检查 Redis 中的会话是否仍然活跃且未过期
	if sessionData.IsActive && time.Since(sessionData.LastHeartbeat) <= heartbeatTTL {
		return fmt.Errorf("用户已在其他位置登录")
	}

	// 会话已过期或不活跃，可以创建新的会话
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

// Heartbeat 处理心跳请求，同时更新 Redis 和数据库
func (s *userSessionService) Heartbeat(ctx context.Context, req *models.HeartbeatRequest) (*models.HeartbeatResponse, error) {
	now := time.Now()
	heartbeatTimeoutSeconds := int32(heartbeatTTL.Seconds())

	// 更新数据库中的心跳时间，并检查是否过期
	err := s.queries.UpdateHeartbeatWithExpireCheck(ctx, db.UpdateHeartbeatWithExpireCheckParams{
		Column1: now,
		Column2: heartbeatTimeoutSeconds,
		ID:      req.SessionID,
	})
	if err != nil {
		return nil, fmt.Errorf("更新数据库心跳失败: %w", err)
	}

	//  从数据库获取更新后的会话状态
	session, err := s.queries.GetUserSessionByID(ctx, req.SessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &models.HeartbeatResponse{
				SessionID: req.SessionID,
				Status:    "invalid",
				Message:   "会话不存在",
			}, nil
		}
		return nil, fmt.Errorf("获取会话信息失败: %w", err)
	}

	// 检查会话是否仍然活跃
	if !session.IsActive.Bool {
		return &models.HeartbeatResponse{
			SessionID:     req.SessionID,
			Status:        "expired",
			LastHeartbeat: session.LastHeartbeat.Time,
			Message:       "会话已过期",
		}, nil
	}

	// 4. 更新 Redis 中的会话数据
	sessionData := models.RedisSessionData{
		ID:            session.ID,
		UserID:        session.UserID,
		StartTime:     session.StartTime.Time,
		LastHeartbeat: now,
		IsActive:      session.IsActive.Bool,
	}

	if err := s.saveSessionToRedis(ctx, sessionData); err != nil {
		// Redis 更新失败不影响整体功能，只记录日志
		fmt.Printf("更新 Redis 会话失败: %v\n", err)
	}

	return &models.HeartbeatResponse{
		SessionID:     req.SessionID,
		Status:        "ok",
		LastHeartbeat: now,
		Message:       "心跳正常",
	}, nil
}

// GetSessionStatus 获取会话状态
func (s *userSessionService) GetSessionStatus(ctx context.Context, sessionID uuid.UUID) (*models.SessionStatusResponse, error) {
	session, err := s.queries.GetUserSessionByID(ctx, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("会话不存在")
		}
		return nil, fmt.Errorf("获取会话信息失败: %w", err)
	}

	// 检查是否过期
	isExpired := false
	if session.IsActive.Bool && session.LastHeartbeat.Valid {
		isExpired = time.Since(session.LastHeartbeat.Time) > heartbeatTTL
	}

	return &models.SessionStatusResponse{
		SessionID:        sessionID,
		IsActive:         session.IsActive.Bool,
		LastHeartbeat:    session.LastHeartbeat.Time,
		IsExpired:        isExpired,
		CanCreateReading: session.IsActive.Bool && !isExpired,
	}, nil
}

// EndUserSession 手动结束会话
func (s *userSessionService) EndUserSession(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now()

	// 更新数据库
	err := s.queries.EndUserSession(ctx, db.EndUserSessionParams{
		EndTime: sql.NullTime{Time: now, Valid: true},
		ID:      sessionID,
	})
	if err != nil {
		return fmt.Errorf("结束数据库会话失败: %w", err)
	}

	// 从 Redis 中删除会话数据
	session, err := s.queries.GetUserSessionByID(ctx, sessionID)
	if err == nil {
		userSessionKey := s.buildUserSessionKey(session.UserID)
		if delErr := s.redisClient.Delete(ctx, userSessionKey); delErr != nil {
			// Redis 删除失败不影响整体功能
			fmt.Printf("从 Redis 删除会话失败: %v\n", delErr)
		}
	}

	return nil
}

// CleanupExpiredSessions 清理过期会话（定时任务使用）
func (s *userSessionService) CleanupExpiredSessions(ctx context.Context) error {
	heartbeatTimeoutSeconds := int32(heartbeatTTL.Seconds())

	err := s.queries.CleanupExpiredSessions(ctx, heartbeatTimeoutSeconds)
	if err != nil {
		return fmt.Errorf("清理过期会话失败: %w", err)
	}

	return nil
}

// GetActiveUserSessionByUserID 根据用户ID获取活跃会话
func (s *userSessionService) GetActiveUserSessionByUserID(ctx context.Context, userID uuid.UUID) (*db.GetActiveUserSessionByUserIDRow, error) {
	row, err := s.queries.GetActiveUserSessionByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}
