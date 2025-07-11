package service

import (
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/models"
	"context"
)

// AuthService 认证服务接口
type AuthService interface {
	// ValidateInviteCode 验证邀请码
	ValidateInviteCode(ctx context.Context, code string) (*models.InviteCode, error)
	// GenerateJWT 生成JWT令牌
	GenerateJWT(userID string, inviteCodeID int) (string, error)
	// ValidateJWT 验证JWT令牌
	ValidateJWT(tokenString string) (*models.JWTClaims, error)
}

// authService 认证服务实现
type authService struct {
	queries *db.Queries
}

// NewAuthService 创建认证服务实例
func NewAuthService(queries *db.Queries) AuthService {
	return &authService{queries: queries}
}

// ValidateInviteCode 实现邀请码验证逻辑
func (s *authService) ValidateInviteCode(ctx context.Context, code string) (*models.InviteCode, error) {
	// TODO: 实现邀请码验证逻辑
	// 1. 查找邀请码
	// 2. 验证是否已被使用
	// 3. 返回邀请码信息
	return nil, nil
}

// GenerateJWT 实现JWT生成逻辑
func (s *authService) GenerateJWT(userID string, inviteCodeID int) (string, error) {
	// TODO: 实现JWT生成逻辑
	// 1. 创建JWT claims
	// 2. 生成并签名token
	return "", nil
}

// ValidateJWT 实现JWT验证逻辑
func (s *authService) ValidateJWT(tokenString string) (*models.JWTClaims, error) {
	// TODO: 实现JWT验证逻辑
	// 1. 解析token
	// 2. 验证签名
	// 3. 返回claims
	return nil, nil
}
