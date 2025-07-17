package service

import (
	"NewsEyeTracking/internal/database"
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// UserSessionService 用户会话服务接口
type UserSessionService interface {
	// CreateOrGetUserSession 创建或获取用户的活跃会话
	CreateOrGetUserSession(ctx context.Context, userID uuid.UUID) (*models.CreateUserSessionResponse, error)
	
	// Heartbeat 处理心跳请求
	Heartbeat(ctx context.Context, req *models.HeartbeatRequest) (*models.HeartbeatResponse, error)
	
	// ProcessDataUpload 处理数据上传（包含心跳分流）
	ProcessDataUpload(ctx context.Context, req *models.DataUploadRequest) (*models.DataUploadResponse, error)
	
	// GetSessionStatus 获取会话状态
	GetSessionStatus(ctx context.Context, sessionID uuid.UUID) (*models.SessionStatusResponse, error)
	
	// EndUserSession 结束用户会话
	EndUserSession(ctx context.Context, sessionID uuid.UUID) error
	
	// CheckSingleSessionLimit 检查单会话限制
	CheckSingleSessionLimit(ctx context.Context, userID uuid.UUID) error
}

// RedisSessionData Redis中存储的会话数据
type RedisSessionData struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	StartTime     time.Time `json:"start_time"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	IsActive      bool      `json:"is_active"`
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
	// Redis键格式，以及存储数据格式
	//session:123e4567-e89b-12d3-a456-426614174000:987fcdeb-51d3-4c8a-9b12-345678901234
	sessionKeyPrefix = "session:"
	userSessionsKey  = "user_sessions:"
	heartbeatTTL     = 5 * time.Minute // 心跳TTL为5分钟
)

// buildSessionKey 构建会话Redis键
func (s *userSessionService) buildSessionKey(userID, sessionID uuid.UUID) string {
	return fmt.Sprintf("%s%s:%s", sessionKeyPrefix, userID.String(), sessionID.String())
}
/*
// buildUserSessionsKey 构建用户会话列表键
func (s *userSessionService) buildUserSessionsKey(userID uuid.UUID) string {
	return fmt.Sprintf("%s%s", userSessionsKey, userID.String())
}*/

// CreateOrGetUserSession 创建或获取用户的活跃会话
func (s *userSessionService) CreateOrGetUserSession(ctx context.Context, userID uuid.UUID) (*models.CreateUserSessionResponse, error) {
	// 首先检查单会话限制
	if err := s.CheckSingleSessionLimit(ctx, userID); err != nil {
		return nil, err
	}
	
	// 检查用户是否已有活跃会话
	pattern := fmt.Sprintf("%s%s:*", sessionKeyPrefix, userID.String())
	
	keys, err := s.redisClient.Keys(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("查询用户会话失败: %v", err)
	}
	
	// 检查现有会话
	for _, key := range keys {
		sessionData, err := s.getSessionFromRedis(ctx, key)
		if err != nil {
			continue
		}
		
		// 检查是否是活跃会话且未过期
		if sessionData.IsActive && time.Since(sessionData.LastHeartbeat) <= heartbeatTTL {
			return &models.CreateUserSessionResponse{
				SessionID:     sessionData.ID,
				UserID:        sessionData.UserID,
				StartTime:     sessionData.StartTime,
				LastHeartbeat: sessionData.LastHeartbeat,
				IsActive:      sessionData.IsActive,
			}, nil
		}
	}
	
	// 创建新会话
	now := time.Now()
	sessionID := uuid.New()
	
	sessionData := RedisSessionData{
		ID:            sessionID,
		UserID:        userID,
		StartTime:     now,
		LastHeartbeat: now,
		IsActive:      true,
	}
	
	// 存储到Redis
	if err := s.saveSessionToRedis(ctx, sessionData); err != nil {
		return nil, fmt.Errorf("保存会话到Redis失败: %v", err)
	}
	
	// 同时在数据库中创建记录（用于数据关联）
	if err := s.createUserSessionInDB(ctx, sessionData); err != nil {
		log.Printf("创建数据库会话记录失败: %v", err)
	}
	
	return &models.CreateUserSessionResponse{
		SessionID:     sessionID,
		UserID:        userID,
		StartTime:     now,
		LastHeartbeat: now,
		IsActive:      true,
	}, nil
}

// Heartbeat 处理心跳请求
func (s *userSessionService) Heartbeat(ctx context.Context, req *models.HeartbeatRequest) (*models.HeartbeatResponse, error) {
	// 从Redis获取会话数据
	sessionData, err := s.getSessionByID(ctx, req.SessionID)
	if err != nil {
		return &models.HeartbeatResponse{
			SessionID: req.SessionID,
			Status:    "invalid",
			Message:   "会话不存在或已过期",
		}, nil
	}
	
	// 更新心跳时间
	sessionData.LastHeartbeat = req.Timestamp
	
	// 保存回Redis并刷新TTL
	if err := s.saveSessionToRedis(ctx, *sessionData); err != nil {
		return &models.HeartbeatResponse{
			SessionID: req.SessionID,
			Status:    "error",
			Message:   "更新心跳失败",
		}, nil
	}
	
	return &models.HeartbeatResponse{
		SessionID:     req.SessionID,
		Status:        "ok",
		LastHeartbeat: req.Timestamp,
	}, nil
}

// ProcessDataUpload 处理数据上传（包含心跳分流）
func (s *userSessionService) ProcessDataUpload(ctx context.Context, req *models.DataUploadRequest) (*models.DataUploadResponse, error) {
	response := &models.DataUploadResponse{
		SessionID: req.SessionID,
	}
	
	// 更新心跳时间
	heartbeatReq := &models.HeartbeatRequest{
		SessionID: req.SessionID,
		Timestamp: req.Timestamp,
	}
	
	heartbeatResp, err := s.Heartbeat(ctx, heartbeatReq)
	if err != nil || heartbeatResp.Status != "ok" {
		return nil, fmt.Errorf("心跳更新失败: %v", err)
	}
	
	// 根据数据类型进行处理
	switch req.DataType {
	case "heartbeat":
		response.ProcessedType = "heartbeat_only"
		response.HeartbeatStatus = "ok"
		response.Message = "心跳数据已处理"
		
	case "eyetracking":
		response.ProcessedType = "eyetracking_only"
		response.HeartbeatStatus = "updated"
		response.DataSize = int64(len(req.CompressedData))
		response.Message = "眼动数据已处理"
		
	case "mixed":
		response.ProcessedType = "mixed"
		response.HeartbeatStatus = "ok"
		response.DataSize = int64(len(req.CompressedData))
		response.Message = "混合数据已处理，心跳数据已分离"
		
		// TODO: 实现混合数据的分离逻辑
		// 这里可以解析压缩数据，分离心跳数据和眼动数据
		
	default:
		return nil, fmt.Errorf("不支持的数据类型: %s", req.DataType)
	}
	
	return response, nil
}

// GetSessionStatus 获取会话状态
func (s *userSessionService) GetSessionStatus(ctx context.Context, sessionID uuid.UUID) (*models.SessionStatusResponse, error) {
	sessionData, err := s.getSessionByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取会话状态失败: %v", err)
	}
	
	// 检查是否过期
	isExpired := time.Since(sessionData.LastHeartbeat) > heartbeatTTL
	
	return &models.SessionStatusResponse{
		SessionID:        sessionData.ID,
		IsActive:         sessionData.IsActive && !isExpired,
		LastHeartbeat:    sessionData.LastHeartbeat,
		IsExpired:        isExpired,
		CanCreateReading: sessionData.IsActive && !isExpired,
	}, nil
}

// EndUserSession 结束用户会话
func (s *userSessionService) EndUserSession(ctx context.Context, sessionID uuid.UUID) error {
	sessionData, err := s.getSessionByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("获取会话失败: %v", err)
	}
	
	// 标记为非活跃
	sessionData.IsActive = false
	
	// 更新Redis
	if err := s.saveSessionToRedis(ctx, *sessionData); err != nil {
		return fmt.Errorf("更新会话状态失败: %v", err)
	}
	
	return nil
}

// CheckSingleSessionLimit 检查单会话限制
func (s *userSessionService) CheckSingleSessionLimit(ctx context.Context, userID uuid.UUID) error {
	// 查找所有活跃会话
	pattern := fmt.Sprintf("%s*", sessionKeyPrefix)
	keys, err := s.redisClient.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("查询活跃会话失败: %v", err)
	}
	
	// 检查其他用户的活跃会话
	for _, key := range keys {
		sessionData, err := s.getSessionFromRedis(ctx, key)
		if err != nil {
			continue
		}
		
		// 跳过当前用户的会话
		if sessionData.UserID == userID {
			continue
		}
		
		// 如果找到其他用户的活跃会话，强制结束
		if sessionData.IsActive && time.Since(sessionData.LastHeartbeat) <= heartbeatTTL {
			sessionData.IsActive = false
			if err := s.saveSessionToRedis(ctx, *sessionData); err != nil {
				log.Printf("强制结束其他用户会话失败: %v", err)
			} else {
				log.Printf("强制结束用户 %s 的会话", sessionData.UserID)
			}
		}
	}
	
	return nil
}

// getSessionByID 根据会话ID获取会话数据
func (s *userSessionService) getSessionByID(ctx context.Context, sessionID uuid.UUID) (*RedisSessionData, error) {
	// 由于我们不知道用户ID，需要搜索所有会话
	pattern := fmt.Sprintf("%s*:%s", sessionKeyPrefix, sessionID.String())
	keys, err := s.redisClient.Keys(ctx, pattern)
	if err != nil {
		return nil, err
	}
	
	if len(keys) == 0 {
		return nil, fmt.Errorf("会话不存在")
	}
	
	return s.getSessionFromRedis(ctx, keys[0])
}

// getSessionFromRedis 从Redis获取会话数据
func (s *userSessionService) getSessionFromRedis(ctx context.Context, key string) (*RedisSessionData, error) {
	data, err := s.redisClient.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("会话不存在")
		}
		return nil, err
	}
	
	var sessionData RedisSessionData
	if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
		return nil, fmt.Errorf("解析会话数据失败: %v", err)
	}
	
	return &sessionData, nil
}

// saveSessionToRedis 保存会话数据到Redis
func (s *userSessionService) saveSessionToRedis(ctx context.Context, sessionData RedisSessionData) error {
	key := s.buildSessionKey(sessionData.UserID, sessionData.ID)
	
	data, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("序列化会话数据失败: %v", err)
	}
	
	return s.redisClient.Set(ctx, key, string(data), heartbeatTTL)
}

// createUserSessionInDB 在数据库中创建会话记录（用于数据关联）
func (s *userSessionService) createUserSessionInDB(ctx context.Context, sessionData RedisSessionData) error {
	// 这里需要等待数据库迁移完成后实现
	// 暂时返回nil
	return nil
}
