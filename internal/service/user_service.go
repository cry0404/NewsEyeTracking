package service

import (
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/models"
	"NewsEyeTracking/internal/api/middleware"

	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)
//这里还缺失了检验邀请码这一逻辑，邀请码应该使用 hash 值来跟数据库中的做对比
// UserService 用户服务接口
type UserService interface {
	// GetUserByID 根据ID获取用户信息
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	// UpdateUser 统一的用户创建/更新接口，支持邀请码创建和用户ID更新
	UpdateUser(ctx context.Context, userID string, req *models.UserRequest) (*models.User, error)
	// GetUserABTestConfig 获取用户A/B测试配置
	GetUserABTestConfig(ctx context.Context, inviteCodeID uuid.UUID) (*models.ABTestConfig, error)
	// UpdateLoginState 更新登录状态
	UpdateLoginState(ctx context.Context, userID uuid.UUID) (string, error)
}

// userService 用户服务实现
type userService struct {
	queries *db.Queries
}

// NewUserService 创建用户服务实例
func NewUserService(queries *db.Queries) UserService {
	return &userService{queries: queries}
}



// GetUserByID 根据ID获取用户信息
func (s *userService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	newUserid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("userid 解析失败，请确认 userid 格式")
	}
	User, err := s.queries.GetUserByID(ctx, newUserid)
	if err != nil {
		return nil, err
	}

	// 转换 db.User 到 models.User, 业务层逻辑和数据库逻辑，保证前端验证的非空性了
	return &models.User{
		ID:                  User.ID,
		//Email:               User.Email,
		Gender:              nullStringToPtr(User.Gender),
		Age:                 nullInt32ToPtr(User.Age),
		DateOfBirth:         nullTimeToPtr(User.DateOfBirth),
		EducationLevel:      nullStringToPtr(User.EducationLevel),
		Residence:           nullStringToPtr(User.Residence),
		WeeklyReadingHours:  nullInt32ToPtr(User.WeeklyReadingHours),
		PrimaryNewsPlatforms: stringToSlice(nullStringToPtr(User.PrimaryNewsPlatform)),
		IsActiveSearcher:    User.IsActiveSearcher.Bool,
		IsColorblind:        User.IsColorblind.Bool,
		VisionStatus:        nullStringToPtr(User.VisionStatus),
		IsVisionCorrected:   User.IsVisionCorrected.Bool,
		CreatedAt:           User.CreatedAt.Time,
	}, nil
}

// UpdateUser 统一的用户更新接口，所有字段都不强制要求
func (s *userService) UpdateUser(ctx context.Context, userID string, req *models.UserRequest) (*models.User, error) {
	
	newUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("用户ID解析失败，请确认用户ID格式: %w", err)
	}

	//这里需要更新 code 中的使用，如果登录成功了，那么邀请码就算使用了
	var dateOfBirth *time.Time
	if req.DateOfBirth != nil {
		parsedDate, err := parseDate(*req.DateOfBirth)
		if err != nil {
			return nil, fmt.Errorf("日期格式错误: %w", err)
		}
		dateOfBirth = &parsedDate
	}

	// 先尝试更新用户
	updateParams := db.UpdateUserParams{
		ID:                  newUserID,
		Gender:              buildNullStringFromPtr(req.Gender),
		Age:                 buildNullInt32FromPtr(req.Age),
		DateOfBirth:         buildNullTimeFromPtr(dateOfBirth),
		EducationLevel:      buildNullStringFromPtr(req.EducationLevel),
		Residence:           buildNullStringFromPtr(req.Residence),
		WeeklyReadingHours:  buildNullInt32FromPtr(req.WeeklyReadingHours),
		PrimaryNewsPlatform: buildNullStringFromSlice(req.PrimaryNewsPlatforms),
		IsActiveSearcher:    buildNullBoolFromPtr(req.IsActiveSearcher),
		IsColorblind:        buildNullBoolFromPtr(req.IsColorblind),
		VisionStatus:        buildNullStringFromPtr(req.VisionStatus),
		IsVisionCorrected:   buildNullBoolFromPtr(req.IsVisionCorrected),
	}

	updatedUser, err := s.queries.UpdateUser(ctx, updateParams)
	if err != nil {
		// 如果用户不存在，尝试创建用户
if err == sql.ErrNoRows {
			// 需要从邀请码获取 email
			inviteInfo, inviteErr := s.queries.GetIdAndEmailByCodeID(ctx, newUserID)
			if inviteErr != nil {
				return nil, fmt.Errorf("用户不存在且无法通过邀请码创建: %w", inviteErr)
			}

			// 创建用户
			createParams := db.CreateUserParams{
				ID:                  newUserID,
				Email:               inviteInfo.Email,
				Gender:              buildNullStringFromPtr(req.Gender),
				Age:                 buildNullInt32FromPtr(req.Age),
				DateOfBirth:         buildNullTimeFromPtr(dateOfBirth),
				EducationLevel:      buildNullStringFromPtr(req.EducationLevel),
				Residence:           buildNullStringFromPtr(req.Residence),
				WeeklyReadingHours:  buildNullInt32FromPtr(req.WeeklyReadingHours),
				PrimaryNewsPlatform: buildNullStringFromSlice(req.PrimaryNewsPlatforms),
				IsActiveSearcher:    buildNullBoolFromPtr(req.IsActiveSearcher),
				IsColorblind:        buildNullBoolFromPtr(req.IsColorblind),
				VisionStatus:        buildNullStringFromPtr(req.VisionStatus),
				IsVisionCorrected:   buildNullBoolFromPtr(req.IsVisionCorrected),
			}

			updatedUser, err = s.queries.CreateUser(ctx, createParams)
			if err != nil {
				return nil, fmt.Errorf("创建用户失败: %w", err)
			}
			return &models.User{ID: updatedUser.ID}, nil
		} else {
			return nil, fmt.Errorf("更新用户信息失败: %w", err)
		}
	}
   	//在最后这个部分更新邀请码已经使用
	
	// 转换为模型对象
	return &models.User{
		ID:                  updatedUser.ID,
		//Email:               updatedUser.Email,
		Gender:              nullStringToPtr(updatedUser.Gender),
		Age:                 nullInt32ToPtr(updatedUser.Age),
		DateOfBirth:         nullTimeToPtr(updatedUser.DateOfBirth),
		EducationLevel:      nullStringToPtr(updatedUser.EducationLevel),
		Residence:           nullStringToPtr(updatedUser.Residence),
		WeeklyReadingHours:  nullInt32ToPtr(updatedUser.WeeklyReadingHours),
		PrimaryNewsPlatforms: stringToSlice(nullStringToPtr(updatedUser.PrimaryNewsPlatform)),
		IsActiveSearcher:    updatedUser.IsActiveSearcher.Bool,
		IsColorblind:        updatedUser.IsColorblind.Bool,
		VisionStatus:        nullStringToPtr(updatedUser.VisionStatus),
		IsVisionCorrected:   updatedUser.IsVisionCorrected.Bool,
		CreatedAt:           updatedUser.CreatedAt.Time,
	}, nil
}



// GetUserABTestConfig 获取用户A/B测试配置
func (s *userService) GetUserABTestConfig(ctx context.Context, inviteCodeID uuid.UUID) (*models.ABTestConfig, error) {
	info, err := s.queries.GetABTestConfigByInviteCodeID(ctx, inviteCodeID)
	if err != nil {
		return nil, fmt.Errorf("获取A/B测试配置失败: %w", err)
	}
	
	return &models.ABTestConfig{
		HasRecommend:       info.HasRecommend.Bool,
		HasMoreInformation: info.HasMoreInformation.Bool,
	}, nil
}

func (s *userService) UpdateLoginState(ctx context.Context, userID uuid.UUID) (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", fmt.Errorf("加载环境变量失败: %w", err)
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT_SECRET 环境变量未设置")
	}

	// JWT 过期时间
	expireDuration :=  60 * time.Minute//暂时先调整这么多，等到上线再调整为 15min？
	token, err := middleware.MakeJWT(userID, jwtSecret, expireDuration)
	if err != nil {
		return "", fmt.Errorf("生成JWT token失败: %w", err)
	}

	
	return token, nil
}
// 辅助函数：解析日期字符串
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// 辅助函数：将 sql.NullString 转换为 *string
func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// 辅助函数：将 sql.NullInt32 转换为 *int
func nullInt32ToPtr(ni sql.NullInt32) *int {
	if ni.Valid {
		val := int(ni.Int32)
		return &val
	}
	return nil
}

// 辅助函数：将 sql.NullTime 转换为 *time.Time
func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// 新的辅助函数：从指针直接构建 sql.Null 类型
func buildNullStringFromPtr(value *string) sql.NullString {
	if value != nil {
		return sql.NullString{String: *value, Valid: true}
	}
	return sql.NullString{Valid: false}
}

func buildNullInt32FromPtr(value *int) sql.NullInt32 {
	if value != nil {
		return sql.NullInt32{Int32: int32(*value), Valid: true}
	}
	return sql.NullInt32{Valid: false}
}

func buildNullTimeFromPtr(value *time.Time) sql.NullTime {
	if value != nil {
		return sql.NullTime{Time: *value, Valid: true}
	}
	return sql.NullTime{Valid: false}
}

func buildNullBoolFromPtr(value *bool) sql.NullBool {
	if value != nil {
		return sql.NullBool{Bool: *value, Valid: true}
	}
	return sql.NullBool{Valid: false}
}

// 新增的转换函数：处理字符串数组与单一字符串之间的转换
// 将字符串数组转换为逗号分隔的字符串（用于存储到数据库）
func buildNullStringFromSlice(values []string) sql.NullString {
	if len(values) > 0 {
		return sql.NullString{String: strings.Join(values, ","), Valid: true}
	}
	return sql.NullString{Valid: false}
}

// 将逗号分隔的字符串转换为字符串数组（用于从数据库读取）
func stringToSlice(value *string) []string {
	if value == nil || *value == "" {
		return []string{}
	}
	return strings.Split(*value, ",")
}

/*
func validateRequiredFields(req *models.RegisterRequest) error {
	if req.Gender == nil {
		return fmt.Errorf("性别字段不能为空")
	}
	if req.Age == nil {
		return fmt.Errorf("年龄字段不能为空")
	}
	if req.DateOfBirth == nil {
		return fmt.Errorf("出生日期字段不能为空")
	}
	if req.EducationLevel == nil {
		return fmt.Errorf("教育水平字段不能为空")
	}
	if req.Residence == nil {
		return fmt.Errorf("居住地字段不能为空")
	}
	if req.WeeklyReadingHours == nil {
		return fmt.Errorf("每周阅读时长字段不能为空")
	}
	if req.PrimaryNewsPlatform == nil {
		return fmt.Errorf("主要新闻平台字段不能为空")
	}
	if req.IsActiveSearcher == nil {
		return fmt.Errorf("是否主动搜索字段不能为空")
	}
	if req.IsColorblind == nil {
		return fmt.Errorf("是否色盲字段不能为空")
	}
	if req.VisionStatus == nil {
		return fmt.Errorf("视力状况字段不能为空")
	}
	if req.IsVisionCorrected == nil {
		return fmt.Errorf("是否视力矫正字段不能为空")
	}
	return nil
}
*/ 
//都可以为空，不需要强制验证了

/*
// CreateUser 实现用户创建逻辑（包含JWT生成）
func (s *userService) CreateUser(ctx context.Context, req *models.UserRequest) (*models.UserRegisterResponse, error) {
//现在创建不需要 invitecode 了，随便更新
	UserInfo, err := s.queries.GetIdAndEmailByCode(ctx, req.InviteCode)
	if err != nil {
		return nil, fmt.Errorf("邀请码无效")
	}


	if err := validateRequiredFields(req); err != nil {
		return nil, fmt.Errorf("字段验证失败")
	}


	var dateOfBirth *time.Time
	if req.DateOfBirth != nil {
		parsedDate, err := parseDate(*req.DateOfBirth)
		if err != nil {
			return nil, fmt.Errorf("日期格式错误")
		}
		dateOfBirth = &parsedDate
	}


	User, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		ID:                  UserInfo.ID,
		Email:               UserInfo.Email,
		Gender:              sql.NullString{String: *req.Gender, Valid: req.Gender != nil},
		Age:                 sql.NullInt32{Int32: int32(*req.Age), Valid: req.Age != nil},
		DateOfBirth:         sql.NullTime{Time: *dateOfBirth, Valid: req.DateOfBirth != nil},
		EducationLevel:      sql.NullString{String: *req.EducationLevel, Valid: req.EducationLevel != nil},
		Residence:           sql.NullString{String: *req.Residence, Valid: req.Residence != nil},
		WeeklyReadingHours:  sql.NullInt32{Int32: int32(*req.WeeklyReadingHours), Valid: req.WeeklyReadingHours != nil},
		PrimaryNewsPlatform: sql.NullString{String: *req.PrimaryNewsPlatform, Valid: req.PrimaryNewsPlatform != nil},
		IsActiveSearcher:    sql.NullBool{Bool: *req.IsActiveSearcher, Valid: req.IsActiveSearcher != nil},
		IsColorblind:        sql.NullBool{Bool: *req.IsColorblind, Valid: req.IsColorblind != nil},
		VisionStatus:        sql.NullString{String: *req.VisionStatus, Valid: req.VisionStatus != nil},
		IsVisionCorrected:   sql.NullBool{Bool: *req.IsVisionCorrected, Valid: req.IsVisionCorrected != nil},
	})

	if err != nil {
		return nil, fmt.Errorf("创建用户失败")
	}

	// 这里标记一次就够了， user_service 调用 handleUser 的方法
	err = s.queries.MarkInviteCodeAsUsed(ctx, req.InviteCode)
	if err != nil {
		return nil, fmt.Errorf("邀请码使用失败")
	}

	//把生成 jwt token 直接封装成一个方法
	err = godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("加载环境变量失败: %w", err)
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET 环境变量未设置")
	}

	// JWT 过期时间
	expireDuration :=  7 * 24 * time.Hour//暂时先调整这么多，等到上线再调整为 15min？
	token, err := middleware.MakeJWT(User.ID, jwtSecret, expireDuration)
	if err != nil {
		return nil, fmt.Errorf("生成JWT token失败: %w", err)
	}


	return &models.UserRegisterResponse{
		UserID: User.ID,
		Token:  token,
	}, nil
}*/