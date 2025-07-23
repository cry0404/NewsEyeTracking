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
	Count   			int64     `json:"count" db:"count"`
}

//现在不需要做任何限定了
type User struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Gender         *string    `json:"gender" db:"gender"`
	Age            *int       `json:"age" db:"age"`
	DateOfBirth    *time.Time `json:"date_of_birth" db:"date_of_birth"`
	EducationLevel *string    `json:"education_level" db:"education_level"`
	Residence      *string    `json:"residence" db:"residence"`

	// 新闻阅读习惯
	WeeklyReadingHours  *int    `json:"weekly_reading_hours" db:"weekly_reading_hours"`
	PrimaryNewsPlatforms []string `json:"primary_news_platforms" db:"primary_news_platforms"`
	IsActiveSearcher    bool    `json:"is_active_searcher" db:"is_active_searcher"`

	// 视觉相关
	IsColorblind      bool    `json:"is_colorblind" db:"is_colorblind"`
	VisionStatus      *string `json:"vision_status" db:"vision_status"`
	IsVisionCorrected bool    `json:"is_vision_corrected" db:"is_vision_corrected"`


	InviteCode *InviteCode `json:"invite_code,omitempty"` // 关联的邀请码信息

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ExperimentConfig 实验配置（从邀请码继承）
type ExperimentConfig struct {
	HasRecommend       bool `json:"has_recommend"`
	HasMoreInformation bool `json:"has_more_information"`
}


/*type UserRegisterRequest struct {
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
}*/


type UserRequest struct {
	// 基本信息
	// Email          *string `json:"email" binding:"omitempty,email,max=255" validate:"omitempty,email"`
	Gender         *string `json:"gender" binding:"omitempty,oneof=男 女" validate:"omitempty,oneof=男 女"`
	Age            *int    `json:"age" binding:"omitempty,min=16,max=100" validate:"omitempty,min=16,max=100"`
	DateOfBirth    *string `json:"date_of_birth" binding:"omitempty" validate:"omitempty,datetime=2006-01-02"`
	EducationLevel *string `json:"education_level" binding:"omitempty,oneof='小学' '初中' '高中' '大专' '本科' '硕士' '博士'" validate:"omitempty,oneof='小学' '初中' '高中' '大专' '本科' '硕士' '博士'"`
	Residence      *string `json:"residence" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100"`

	// 新闻阅读习惯
	WeeklyReadingHours  *int    `json:"weekly_reading_hours" binding:"omitempty,oneof=1 2 3 4 5 6 7" validate:"omitempty,oneof=1 2 3 4 5 6 7"`
	PrimaryNewsPlatforms []string `json:"primary_news_platforms" binding:"omitempty" validate:"omitempty"`
	IsActiveSearcher    *bool   `json:"is_active_searcher" binding:"omitempty" validate:"omitempty"`

	// 视觉相关
	IsColorblind      *bool   `json:"is_colorblind" binding:"omitempty" validate:"omitempty"`
	VisionStatus      *string `json:"vision_status" binding:"omitempty,oneof='远视' '近视' '无'" validate:"omitempty,oneof='远视' '近视' '无'"`
	IsVisionCorrected *bool   `json:"is_vision_corrected" binding:"omitempty" validate:"omitempty"`
}
/*
// UserRegisterResponse 用户注册响应
type UserRegisterResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Token  string    `json:"token"`
}*/

// UserProfileResponse 用户信息响应
type UserProfileResponse struct {
	ID                  uuid.UUID        `json:"id"`
	Email               string           `json:"email"`
	Gender              *string          `json:"gender"`
	Age                 *int             `json:"age"`
	EducationLevel      *string          `json:"education_level"`
	Residence           *string          `json:"residence"`
	WeeklyReadingHours  *int             `json:"weekly_reading_hours"`
	PrimaryNewsPlatforms []string         `json:"primary_news_platforms"`
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
	EducationPrimary     = "小学"
	EducationMiddle      = "初中"
	EducationHighSchool  = "高中"
	EducationJuniorCollege = "大专"
	EducationBachelor    = "本科"
	EducationMaster      = "硕士"
	EducationDoctor      = "博士"

	// 每周阅读时长映射
	ReadingHours1  = 1 // 1小时
	ReadingHours3  = 2 // 3小时
	ReadingHours5  = 3 // 5小时
	ReadingHours10 = 4 // 10小时
	ReadingHours20 = 5 // 20小时
	ReadingHours30 = 6 // 30小时
	ReadingHours40AndUp = 7 // 40小时及以上

	// 新闻平台
	PlatformToutiao     = "今日头条"
	PlatformTencentNews = "腾讯新闻"
	PlatformSinaNews    = "新浪新闻"
	PlatformSohuNews    = "搜狐新闻"
	PlatformNeteaseNews = "网易新闻"
	PlatformPhoenixNews = "凤凰新闻"
	PlatformPeoplesDaily = "人民日报"
	PlatformCCTVNews     = "央视新闻"
	PlatformXinhuaNet    = "新华网"
	PlatformGuangmingDaily = "光明日报"

	// 视力状况
	VisionFarsighted  = "远视"
	VisionNearsighted = "近视"
	VisionNormal      = "无"
)

// GetReadingHoursText 根据数值获取阅读时长文本
func GetReadingHoursText(hours int) string {
	switch hours {
	case ReadingHours1:
		return "1小时"
	case ReadingHours3:
		return "3小时"
	case ReadingHours5:
		return "5小时"
	case ReadingHours10:
		return "10小时"
	case ReadingHours20:
		return "20小时"
	case ReadingHours30:
		return "30小时"
	case ReadingHours40AndUp:
		return "40小时及以上"
	default:
		return "未知"
	}
}

// ValidateReadingHours 验证阅读时长是否有效
func ValidateReadingHours(hours int) bool {
	return hours >= ReadingHours1 && hours <= ReadingHours40AndUp
}

// RegisterRequest 用户注册请求的别名，保持向后兼容
type RegisterRequest = UserRequest


type ABTestConfig = ExperimentConfig

// JWTClaims JWT令牌声明
type JWTClaims struct {
	UserID       string `json:"user_id"` // UUID 中作为 string 来传输，截获后再做解码
	HasRecommend bool   `json:"has_recommend"`
	HasMoreInfo  bool   `json:"has_more_info"`
	Exp          int64  `json:"exp"`
	Iat          int64  `json:"iat"`
}
