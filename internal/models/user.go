package models

import (
	"time"
)

// InviteCode 邀请码表模型
type InviteCode struct {
	ID                 int       `json:"id" db:"id"`
	Code               string    `json:"code" db:"code"`
	IsUsed             bool      `json:"is_used" db:"is_used"`
	UsedByUserID       *int      `json:"used_by_user_id" db:"used_by_user_id"`
	HasRecommend       bool      `json:"has_recommend" db:"has_recommend"`               // A/B测试：是否启用推荐算法
	HasMoreInformation bool      `json:"has_more_information" db:"has_more_information"` // A/B测试：是否显示更多信息
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

// User 用户表模型
type User struct {
	ID             int        `json:"id" db:"id"`
	Email          string     `json:"email" db:"email"`
	Gender         *string    `json:"gender" db:"gender"`
	Age            *int       `json:"age" db:"age"`
	DateOfBirth    *time.Time `json:"date_of_birth" db:"date_of_birth"`
	EducationLevel *string    `json:"education_level" db:"education_level"`
	Residence      *string    `json:"residence" db:"residence"`

	// 新闻阅读习惯
	WeeklyReadingHours  *int    `json:"weekly_reading_hours" db:"weekly_reading_hours"`
	PrimaryNewsPlatform *string `json:"primary_news_platform" db:"primary_news_platform"`
	IsActiveSearcher    bool    `json:"is_active_searcher" db:"is_active_searcher"`

	// 视觉相关
	IsColorblind      bool    `json:"is_colorblind" db:"is_colorblind"`
	VisionStatus      *string `json:"vision_status" db:"vision_status"`
	IsVisionCorrected bool    `json:"is_vision_corrected" db:"is_vision_corrected"`

	// 实验相关
	InviteCodeID int         `json:"invite_code_id" db:"invite_code_id"`
	InviteCode   *InviteCode `json:"invite_code,omitempty"` // 关联的邀请码信息

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ExperimentConfig 实验配置（从邀请码继承）
type ExperimentConfig struct {
	HasRecommend       bool `json:"has_recommend"`
	HasMoreInformation bool `json:"has_more_information"`
}

// UserRegisterRequest 用户注册请求
type UserRegisterRequest struct {
	InviteCode          string  `json:"invite_code" binding:"required"`
	Email               string  `json:"email" binding:"required,email"`
	Gender              *string `json:"gender"`
	Age                 *int    `json:"age"`
	DateOfBirth         *string `json:"date_of_birth"` // 前端传送字符串格式，后端解析
	EducationLevel      *string `json:"education_level"`
	Residence           *string `json:"residence"`
	WeeklyReadingHours  *int    `json:"weekly_reading_hours"`
	PrimaryNewsPlatform *string `json:"primary_news_platform"`
	IsActiveSearcher    *bool   `json:"is_active_searcher"`
	IsColorblind        *bool   `json:"is_colorblind"`
	VisionStatus        *string `json:"vision_status"`
	IsVisionCorrected   *bool   `json:"is_vision_corrected"`
}

// UserRegisterResponse 用户注册响应
type UserRegisterResponse struct {
	UserID             int    `json:"user_id"`
	Email              string `json:"email"`
	HasRecommend       bool   `json:"has_recommend"`
	HasMoreInformation bool   `json:"has_more_information"`
	Token              string `json:"token"`
}

// UserProfileResponse 用户信息响应
type UserProfileResponse struct {
	ID                  int              `json:"id"`
	Email               string           `json:"email"`
	Gender              *string          `json:"gender"`
	Age                 *int             `json:"age"`
	EducationLevel      *string          `json:"education_level"`
	Residence           *string          `json:"residence"`
	WeeklyReadingHours  *int             `json:"weekly_reading_hours"`
	PrimaryNewsPlatform *string          `json:"primary_news_platform"`
	IsActiveSearcher    bool             `json:"is_active_searcher"`
	ExperimentConfig    ExperimentConfig `json:"experiment_config"`
	CreatedAt           time.Time        `json:"created_at"`
}

// UserUpdateRequest 用户信息更新请求
type UserUpdateRequest struct {
	Email               *string `json:"email"`
	Gender              *string `json:"gender"`
	Age                 *int    `json:"age"`
	DateOfBirth         *string `json:"date_of_birth"`
	EducationLevel      *string `json:"education_level"`
	Residence           *string `json:"residence"`
	WeeklyReadingHours  *int    `json:"weekly_reading_hours"`
	PrimaryNewsPlatform *string `json:"primary_news_platform"`
	IsActiveSearcher    *bool   `json:"is_active_searcher"`
	IsColorblind        *bool   `json:"is_colorblind"`
	VisionStatus        *string `json:"vision_status"`
	IsVisionCorrected   *bool   `json:"is_vision_corrected"`
}

// RegisterRequest 用户注册请求的别名，保持向后兼容
type RegisterRequest = UserRegisterRequest

// ABTestConfig A/B测试配置的别名，保持向后兼容
type ABTestConfig = ExperimentConfig

// JWTClaims JWT令牌声明
type JWTClaims struct {
	UserID       string `json:"user_id"`
	InviteCodeID int    `json:"invite_code_id"`
	HasRecommend bool   `json:"has_recommend"`
	HasMoreInfo  bool   `json:"has_more_info"`
	Exp          int64  `json:"exp"`
	Iat          int64  `json:"iat"`
}
