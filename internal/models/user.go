package models

import (
	"time"

	"github.com/google/uuid"
)

// InviteCode 邀请码表模型
type InviteCode struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	Code               string     `json:"code" db:"code"`
	IsUsed             bool       `json:"is_used" db:"is_used"`
	UsedByUserID       *uuid.UUID `json:"used_by_user_id" db:"used_by_user_id"`
	HasRecommend       bool       `json:"has_recommend" db:"has_recommend"`               // 是否启用推荐算法
	HasMoreInformation bool       `json:"has_more_information" db:"has_more_information"` // 信息
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

// User 表这里应该需要绑定 bind ，应该要做一些 require， 对于前端发送过来的字段
type User struct {
	ID             uuid.UUID  `json:"id" db:"id"`
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

	// 实验相关 - 注意：User.ID 就是来自 invite_codes.id，所以不需要单独的 InviteCodeID
	// 类似于 feed 和 feedid 之间的关系
	InviteCode *InviteCode `json:"invite_code,omitempty"` // 关联的邀请码信息

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ExperimentConfig 实验配置（从邀请码继承）
type ExperimentConfig struct {
	HasRecommend       bool `json:"has_recommend"`
	HasMoreInformation bool `json:"has_more_information"`
}


type UserRegisterRequest struct {
	// 必填字段
	InviteCode string `json:"invite_code" binding:"required,min=1,max=50" validate:"required"`
	//注册时没有 email 字段 Email      string `json:"email" binding:"required,email,max=255" validate:"required,email"`

	// 基本信息（严格验证枚举值）
	Gender         *string `json:"gender" binding:"required,oneof=男 女" validate:"required,oneof=男 女"`
	Age            *int    `json:"age" binding:"required,min=16,max=100" validate:"required,min=16,max=100"`
	DateOfBirth    *string `json:"date_of_birth" binding:"required" validate:"required,datetime=2006-01-02"` // 前端传送字符串格式
	EducationLevel *string `json:"education_level" binding:"required,oneof='高中及以下' '本科' '硕士' '博士及以上'" validate:"required,oneof='高中及以下' '本科' '硕士' '博士及以上'"`
	Residence      *string `json:"residence" binding:"required,min=1,max=100" validate:"required,min=1,max=100"`

	// 新闻阅读习惯
	WeeklyReadingHours  *int    `json:"weekly_reading_hours" binding:"required,oneof=1 2 3 4" validate:"required,oneof=1 2 3 4"` // 1=10小时以下, 2=10-20小时, 3=30小时, 4=30小时及以上
	PrimaryNewsPlatform *string `json:"primary_news_platform" binding:"required,oneof='微信新闻' '今日头条' '新浪微博'" validate:"required,oneof='微信新闻' '今日头条' '新浪微博'"`
	IsActiveSearcher    *bool   `json:"is_active_searcher" binding:"required" validate:"required"`

	// 视觉相关
	IsColorblind      *bool   `json:"is_colorblind" binding:"required" validate:"required"`
	VisionStatus      *string `json:"vision_status" binding:"required,oneof='远视' '近视' '无'" validate:"required,oneof='远视' '近视' '无'"`
	IsVisionCorrected *bool   `json:"is_vision_corrected" binding:"required" validate:"required"`
}


type UserUpdateRequest struct {
	// 基本信息
	// Email          *string `json:"email" binding:"omitempty,email,max=255" validate:"omitempty,email"`
	Gender         *string `json:"gender" binding:"omitempty,oneof=男 女" validate:"omitempty,oneof=男 女"`
	Age            *int    `json:"age" binding:"omitempty,min=16,max=100" validate:"omitempty,min=16,max=100"`
	DateOfBirth    *string `json:"date_of_birth" binding:"omitempty" validate:"omitempty,datetime=2006-01-02"`
	EducationLevel *string `json:"education_level" binding:"omitempty,oneof='高中及以下' '本科' '硕士' '博士及以上'" validate:"omitempty,oneof='高中及以下' '本科' '硕士' '博士及以上'"`
	Residence      *string `json:"residence" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100"`

	// 新闻阅读习惯
	WeeklyReadingHours  *int    `json:"weekly_reading_hours" binding:"omitempty,oneof=1 2 3 4" validate:"omitempty,oneof=1 2 3 4"`
	PrimaryNewsPlatform *string `json:"primary_news_platform" binding:"omitempty,oneof='微信新闻' '今日头条' '新浪微博'" validate:"omitempty,oneof='微信新闻' '今日头条' '新浪微博'"`
	IsActiveSearcher    *bool   `json:"is_active_searcher" binding:"omitempty" validate:"omitempty"`

	// 视觉相关
	IsColorblind      *bool   `json:"is_colorblind" binding:"omitempty" validate:"omitempty"`
	VisionStatus      *string `json:"vision_status" binding:"omitempty,oneof='远视' '近视' '无'" validate:"omitempty,oneof='远视' '近视' '无'"`
	IsVisionCorrected *bool   `json:"is_vision_corrected" binding:"omitempty" validate:"omitempty"`
}

// UserRegisterResponse 用户注册响应
type UserRegisterResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Token  string    `json:"token"`
}

// UserProfileResponse 用户信息响应
type UserProfileResponse struct {
	ID                  uuid.UUID        `json:"id"`
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

// 常量定义，用于前端表单枚举值映射
const (
	// 性别
	GenderMale   = "男"
	GenderFemale = "女"

	// 学历
	EducationHighSchoolOrBelow = "高中及以下"
	EducationBachelor          = "本科"
	EducationMaster            = "硕士"
	EducationDoctorOrAbove     = "博士及以上"

	// 每周阅读时长映射
	ReadingHoursBelow10 = 1 // 10小时以下
	ReadingHours10To20  = 2 // 10-20小时
	ReadingHours30      = 3 // 30小时
	ReadingHours30AndUp = 4 // 30小时及以上

	// 新闻平台
	PlatformWeChat  = "微信新闻"
	PlatformToutiao = "今日头条"
	PlatformWeibo   = "新浪微博"

	// 视力状况
	VisionFarsighted  = "远视"
	VisionNearsighted = "近视"
	VisionNormal      = "无"
)

// GetReadingHoursText 根据数值获取阅读时长文本
func GetReadingHoursText(hours int) string {
	switch hours {
	case ReadingHoursBelow10:
		return "10小时以下"
	case ReadingHours10To20:
		return "10-20小时"
	case ReadingHours30:
		return "30小时"
	case ReadingHours30AndUp:
		return "30小时及以上"
	default:
		return "未知"
	}
}

// ValidateReadingHours 验证阅读时长是否有效
func ValidateReadingHours(hours int) bool {
	return hours >= ReadingHoursBelow10 && hours <= ReadingHours30AndUp
}

// RegisterRequest 用户注册请求的别名，保持向后兼容
type RegisterRequest = UserRegisterRequest

// ABTestConfig A/B测试配置的别名，保持向后兼容
type ABTestConfig = ExperimentConfig

// JWTClaims JWT令牌声明
type JWTClaims struct {
	UserID       string `json:"user_id"` // UUID 中作为 string 来传输，截获后再做解码
	HasRecommend bool   `json:"has_recommend"`
	HasMoreInfo  bool   `json:"has_more_info"`
	Exp          int64  `json:"exp"`
	Iat          int64  `json:"iat"`
}
