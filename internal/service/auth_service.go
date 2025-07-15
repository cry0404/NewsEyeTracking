package service

import (
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/models"
	"context"
	"fmt"
)

// AuthService 认证服务接口， 服务由接口构成
type AuthService interface {
	// ValidateInviteCode 验证邀请码
	ValidateInviteCode(ctx context.Context, code string) (*models.InviteCode, error)
	//认证层目前就考虑验证邀请码
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
	// 查询邀请码 - SQL 已经过滤了已使用的邀请码
	codeInfo, err := s.queries.ValidateInviteCode(ctx, code)

	if err != nil {
		return nil, fmt.Errorf("code 查询失败，邀请码错误 : %w", err)
	}

	// 不需要再次检查 IsUsed，因为 SQL 查询已经过滤了已使用的邀请码
	// 将数据库模型转换为业务模型
	inviteCode := &models.InviteCode{
		ID:                 codeInfo.ID,
		Code:               codeInfo.Code,
		IsUsed:             codeInfo.IsUsed.Bool, 
		HasRecommend:       codeInfo.HasRecommend.Bool,
		HasMoreInformation: codeInfo.HasMoreInformation.Bool,
		CreatedAt:          codeInfo.CreatedAt.Time,
	}

	return inviteCode, nil
}
