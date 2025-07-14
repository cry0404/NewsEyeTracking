package service

import (
	"NewsEyeTracking/internal/db"
	"database/sql"
)

// 合理的架构设计？ service 包含每个所有的service 接口, 通过 service 来调用相应的接口
type Services struct {
	User    UserService
	News    NewsService
	Session SessionService
	Auth    AuthService
}

// NewServices 创建服务层实例
func NewServices(database *sql.DB) *Services {
	queries := db.New(database)

	return &Services{
		User:    NewUserService(queries),
		News:    NewNewsService(queries),
		Session: NewSessionService(queries),
		Auth:    NewAuthService(queries),
	}
}
