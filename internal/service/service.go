package service

import (
	"NewsEyeTracking/internal/db"
	"database/sql"
)

// Services 服务层结构体，包含所有业务服务
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
