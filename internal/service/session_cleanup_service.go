package service

import (
	"NewsEyeTracking/internal/db"
	"context"
	"fmt"
	"log"
	"time"

)

// SessionCleanupService 会话清理服务
type SessionCleanupService struct {
	queries            *db.Queries
	userSessionService UserSessionService
	readingService     SessionService
	cleanupInterval    time.Duration
	sessionTimeout     time.Duration
	stopChan          chan struct{}
}

// NewSessionCleanupService 会话清理服务
func NewSessionCleanupService(
	queries *db.Queries,
	userSessionService UserSessionService,
	readingService SessionService,
) *SessionCleanupService {
	return &SessionCleanupService{
		queries:            queries,
		userSessionService: userSessionService,
		readingService:     readingService,
		cleanupInterval:    30 * time.Second, // 每30秒检查一次
		sessionTimeout:     2 * time.Minute,  // 2分钟无心跳则清理
		stopChan:          make(chan struct{}),
	}
}

// Start 启动清理服务
func (s *SessionCleanupService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	log.Printf("会话清理服务已启动，检查间隔: %v, 会话超时: %v", s.cleanupInterval, s.sessionTimeout)

	for {
		select {
		case <-ticker.C:
			if err := s.cleanupExpiredSessions(ctx); err != nil {
				log.Printf("清理过期会话失败: %v", err)
			}
		case <-s.stopChan:
			log.Println("会话清理服务已停止")
			return
		case <-ctx.Done():
			log.Println("会话清理服务因上下文取消而停止")
			return
		}
	}
}

// Stop 停止清理服务
func (s *SessionCleanupService) Stop() {
	close(s.stopChan)
}

// cleanupExpiredSessions 清理过期会话
func (s *SessionCleanupService) cleanupExpiredSessions(ctx context.Context) error {
	timeoutSeconds := int32(s.sessionTimeout.Seconds())
	

	expiredSessions, err := s.userSessionService.GetExpiredUserSessions(ctx, timeoutSeconds)
	if err != nil {
		return fmt.Errorf("获取过期用户会话失败: %w", err)
	}

	cleanedUserSessions := 0
	cleanedReadingSessions := 0

	for _, session := range expiredSessions {
		log.Printf("开始清理过期用户会话: %s (用户: %s, 最后心跳: %v)", 
			session.ID, session.UserID, session.LastHeartbeat.Time)


		activeSessions, err := s.readingService.GetUserActiveSessions(ctx, session.UserID)
		if err != nil {
			log.Printf("获取用户 %s 的活跃阅读会话失败: %v", session.UserID, err)
		} else {
			for _, readingSession := range activeSessions {
				if err := s.readingService.ForceEndSession(ctx, readingSession.ID); err != nil {
					log.Printf("强制结束阅读会话 %s 失败: %v", readingSession.ID, err)
				} else {
					cleanedReadingSessions++
				}
			}
		}


		if err := s.userSessionService.EndUserSession(ctx, session.ID); err != nil {
			log.Printf("结束用户会话 %s 失败: %v", session.ID, err)
		} else {
			cleanedUserSessions++
		}
	}

	if cleanedUserSessions > 0 || cleanedReadingSessions > 0 {
		log.Printf("清理完成: 用户会话 %d 个, 阅读会话 %d 个", 
			cleanedUserSessions, cleanedReadingSessions)
	}

	return nil
}

// SetCleanupInterval 设置清理间隔
func (s *SessionCleanupService) SetCleanupInterval(interval time.Duration) {
	s.cleanupInterval = interval
}

// SetSessionTimeout 设置会话超时时间
func (s *SessionCleanupService) SetSessionTimeout(timeout time.Duration) {
	s.sessionTimeout = timeout
}
