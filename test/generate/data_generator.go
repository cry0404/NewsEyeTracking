package main

import (
	"NewsEyeTracking/internal/database"
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// 生成虚拟数据的程序
func main() {
	// 加载环境变量
	err := godotenv.Load("../../.env")
	if err != nil {
		fmt.Println("环境变量文件未找到，将使用系统环境变量")
	}

	// 获取数据库连接字符串
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		fmt.Printf("错误: DB_URL环境变量未设置\n")
		return
	}

	// 连接数据库
	database, err := database.Connect(dbURL)
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		return
	}
	defer database.Close()

	ctx := context.Background()

	fmt.Println("开始生成虚拟数据...")

	// 生成10对invite_codes和users
	for i := 1; i <= 10; i++ {
		// 生成一个UUID作为id
		id := uuid.New()

		// 生成邀请码数据
		inviteCode := generateInviteCode(id, i)

		// 生成用户数据
		user := generateUser(id, i)

		// 保存到数据库
		if err := saveData(ctx, database, inviteCode, user); err != nil {
			fmt.Printf("保存第%d组数据失败: %v\n", i, err)
			continue
		}

		fmt.Printf("✓ 成功生成第%d组数据 - 邀请码: %s, 用户: %s\n", i, inviteCode.Code, user.Email)
	}

	fmt.Println("数据生成完成！")
}

// 生成邀请码数据
func generateInviteCode(id uuid.UUID, index int) InviteCodeData {
	// 为每个调用设置不同的种子
	rand.Seed(time.Now().UnixNano() + int64(index*1000))

	return InviteCodeData{
		ID:                 id,
		Code:               fmt.Sprintf("TEST_%04d_%s", index, generateRandomString(6)),
		IsUsed:             false,
		HasRecommend:       rand.Float32() < 0.5, // 50%概率启用推荐算法
		HasMoreInformation: rand.Float32() < 0.5, // 50%概率显示更多信息
		CreatedAt:          time.Now(),
	}
}

// 生成用户数据
func generateUser(id uuid.UUID, index int) UserData {
	// 为每个调用设置不同的种子
	rand.Seed(time.Now().UnixNano() + int64(index*2000))

	// 随机数据集
	genders := []string{"男", "女", "其他"}
	educationLevels := []string{"高中", "本科", "硕士", "博士"}
	residences := []string{"北京", "上海", "广州", "深圳", "杭州", "武汉", "成都", "西安"}
	newsPlatforms := []string{"微信", "微博", "今日头条", "知乎", "腾讯新闻", "网易新闻"}
	visionStatuses := []string{"正常", "近视", "远视"}

	age := 18 + rand.Intn(45) // 18-62岁
	dateOfBirth := time.Date(2024-age, time.Month(rand.Intn(12)+1), rand.Intn(28)+1, 0, 0, 0, 0, time.UTC)

	return UserData{
		ID:                  id,
		Email:               fmt.Sprintf("test_user_%04d@example.com", index),
		Gender:              genders[rand.Intn(len(genders))],
		Age:                 age,
		DateOfBirth:         dateOfBirth,
		EducationLevel:      educationLevels[rand.Intn(len(educationLevels))],
		Residence:           residences[rand.Intn(len(residences))],
		WeeklyReadingHours:  rand.Intn(30) + 5, // 5-34小时
		PrimaryNewsPlatform: newsPlatforms[rand.Intn(len(newsPlatforms))],
		IsActiveSearcher:    rand.Float32() < 0.6,  // 60%是主动搜索者
		IsColorblind:        rand.Float32() < 0.08, // 8%色盲概率
		VisionStatus:        visionStatuses[rand.Intn(len(visionStatuses))],
		IsVisionCorrected:   rand.Float32() < 0.7, // 70%视力矫正
		CreatedAt:           time.Now(),
	}
}

// 保存数据到数据库
func saveData(ctx context.Context, database *sql.DB, inviteCode InviteCodeData, user UserData) error {
	// 开启事务
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	// 插入邀请码数据
	if err := insertInviteCode(ctx, tx, inviteCode); err != nil {
		return fmt.Errorf("插入邀请码失败: %w", err)
	}

	// 插入用户数据
	if err := insertUser(ctx, tx, user); err != nil {
		return fmt.Errorf("插入用户失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// 插入邀请码
func insertInviteCode(ctx context.Context, tx *sql.Tx, data InviteCodeData) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO invite_codes (id, code, is_used, has_recommend, has_more_information, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, data.ID, data.Code, data.IsUsed, data.HasRecommend, data.HasMoreInformation, data.CreatedAt)

	return err
}

// 插入用户
func insertUser(ctx context.Context, tx *sql.Tx, data UserData) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO users (
			id, email, gender, age, date_of_birth, education_level, residence,
			weekly_reading_hours, primary_news_platform, is_active_searcher,
			is_colorblind, vision_status, is_vision_corrected, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, data.ID, data.Email, data.Gender, data.Age, data.DateOfBirth,
		data.EducationLevel, data.Residence, data.WeeklyReadingHours,
		data.PrimaryNewsPlatform, data.IsActiveSearcher, data.IsColorblind,
		data.VisionStatus, data.IsVisionCorrected, data.CreatedAt)

	return err
}

// 生成随机字符串
func generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// 数据结构定义
type InviteCodeData struct {
	ID                 uuid.UUID
	Code               string
	IsUsed             bool
	HasRecommend       bool
	HasMoreInformation bool
	CreatedAt          time.Time
}

type UserData struct {
	ID                  uuid.UUID
	Email               string
	Gender              string
	Age                 int
	DateOfBirth         time.Time
	EducationLevel      string
	Residence           string
	WeeklyReadingHours  int
	PrimaryNewsPlatform string
	IsActiveSearcher    bool
	IsColorblind        bool
	VisionStatus        string
	IsVisionCorrected   bool
	CreatedAt           time.Time
}
