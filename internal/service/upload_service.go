package service

import (
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/utils"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"log"
)

// UploadService 上传服务接口
type UploadService interface {
	// StartMonitoring 启动文件监控和自动上传
	StartMonitoring(ctx context.Context) error
	// CleanupOldData 清理过期数据
	CleanupOldData(ctx context.Context, daysToKeep int) error
	// ForceUpload 强制上传所有文件
	ForceUpload(ctx context.Context) error
}

// uploadService 上传服务实现
type uploadService struct {
	queries  *db.Queries
	uploader *utils.FileUploader
	config   utils.Config
}

// NewUploadService 创建上传服务实例
func NewUploadService(queries *db.Queries) UploadService {
	// 从环境变量读取配置
	maxFiles, _ := strconv.Atoi(os.Getenv("UPLOAD_MAX_FILES"))
	maxSize, _ := strconv.ParseInt(os.Getenv("UPLOAD_MAX_SIZE"), 10, 64)
	checkInterval, _ := time.ParseDuration(os.Getenv("UPLOAD_CHECK_INTERVAL"))

	config := utils.Config{
		TrackingDir:   os.Getenv("UPLOAD_TRACKING_DIR"),
		NewsDir:       os.Getenv("UPLOAD_NEWS_DIR"),
		UploadDir:     os.Getenv("UPLOAD_TEMP_DIR"),
		MaxFiles:      maxFiles,
		MaxSize:       maxSize,
		CheckInterval: checkInterval,
	}

	uploader := utils.NewUploader(config)

	return &uploadService{
		queries:  queries,
		uploader: uploader,
		config:   config,
	}
}

// StartMonitoring 启动文件监控和自动上传
func (s *uploadService) StartMonitoring(ctx context.Context) error {
	log.Println("启动文件监控服务...")

	// 在独立的 goroutine 中运行监控服务
	go func() {
		if err := s.uploader.Start(ctx); err != nil {
			log.Printf("文件监控服务错误: %v", err)
		}
	}()

	// 启动每日定时上传
	go s.startDailyUpload(ctx)

	return nil
}

// CleanupOldData 清理过期数据
func (s *uploadService) CleanupOldData(ctx context.Context, daysToKeep int) error {
	log.Printf("清理 %d 天前的数据...", daysToKeep)

	//这里需要删除一些比较旧的日志记录

	return nil
}

// startDailyUpload 启动每日定时上传任务
func (s *uploadService) startDailyUpload(ctx context.Context) {
	log.Println("启动每日定时上传任务...")

	// 计算下一次凌晨12点的时间
	nextUploadTime := s.calculateNextMidnight()
	log.Printf("下一次定时上传时间: %s", nextUploadTime.Format("2006-01-02 15:04:05"))

	timer := time.NewTimer(time.Until(nextUploadTime))
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("停止每日定时上传任务")
			return
		case <-timer.C:
			log.Println("开始执行每日定时上传...")

			// 执行强制上传
			if err := s.ForceUpload(ctx); err != nil {
				log.Printf("每日定时上传失败: %v", err)
			} else {
				log.Println("每日定时上传完成")
			}

			// 设置下一次上传时间（24小时后）
			nextUploadTime = s.calculateNextMidnight()
			log.Printf("下一次定时上传时间: %s", nextUploadTime.Format("2006-01-02 15:04:05"))
			timer.Reset(time.Until(nextUploadTime))
		}
	}
}

// calculateNextMidnight 计算下一次凌晨12点的时间
func (s *uploadService) calculateNextMidnight() time.Time {
	now := time.Now()
	// 获取今天的凌晨12点
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 如果现在已经过了今天的凌晨12点，就设置为明天的凌晨12点
	if now.After(today) {
		return today.Add(24 * time.Hour)
	}

	return today
}

// ForceUpload 强制上传所有文件
func (s *uploadService) ForceUpload(ctx context.Context) error {
	log.Println("开始强制上传...")

	// 获取当前统计信息
	stats, err := s.uploader.GetStats()
	if err != nil {
		return fmt.Errorf("获取文件统计信息失败: %v", err)
	}

	log.Printf("当前待上传文件数: %d, 总大小: %d bytes",
		stats["total_files"], stats["total_size"])

	// 执行强制上传
	if err := s.uploader.ForceUploadAll(); err != nil {
		return fmt.Errorf("强制上传失败: %v", err)
	}

	log.Println("强制上传完成")
	return nil
}
