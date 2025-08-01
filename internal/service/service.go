package service

import (
	"NewsEyeTracking/internal/database"
	"NewsEyeTracking/internal/db"
	"database/sql"
)

// 合理的架构设计？ service 包含每个所有的service 接口, 通过 service 来调用相应的接口
type Services struct {
	User           UserService
	News           NewsService
	Session        SessionService
	UserSession    UserSessionService
	Auth           AuthService
	Upload         UploadService
	Recommend      *RecommendService      // 推荐服务
	SessionCleanup *SessionCleanupService // 会话清理服务
}

// NewServices 创建服务层实例
func NewServices(database *sql.DB, redisClient *database.RedisClient) *Services {
	queries := db.New(database)

	recommendService := NewRecommendService() // 使用默认地址

	// 初始化各个服务
	userSessionService := NewUserSessionService(queries, redisClient)
	sessionService := NewSessionService(queries)

	// 创建会话清理服务
	sessionCleanupService := NewSessionCleanupService(queries, userSessionService, sessionService)

	return &Services{
		User:           NewUserService(queries),
		News:           NewNewsService(queries, recommendService), // 传递推荐服务
		Session:        sessionService,
		UserSession:    userSessionService,
		Auth:           NewAuthService(queries),
		Upload:         NewUploadService(queries),
		Recommend:      recommendService,
		SessionCleanup: sessionCleanupService,
	}
}
