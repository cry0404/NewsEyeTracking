package service

import (
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/utils"
	"context"
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
}

// uploadService 上传服务实现
type uploadService struct {
	queries      *db.Queries
	uploader     *utils.FileUploader
	config       utils.Config
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
	
	return nil
}

// CleanupOldData 清理过期数据
func (s *uploadService) CleanupOldData(ctx context.Context, daysToKeep int) error {
	log.Printf("清理 %d 天前的数据...", daysToKeep)
	
	//这里需要删除一些比较旧的日志记录
	
	return nil
}

//但是实现顺序可以放在最后？
// UploadFiles 手动触发文件上传， 连接一个 api 来处理手动上传，类似于 webhook？
/*func (s *uploadService) UploadFiles(ctx context.Context) error {
	log.Println("手动触发文件上传...")
	
	// 获取当前统计信息
	stats, err := s.uploader.GetStats()
	if err != nil {
		return fmt.Errorf("获取文件统计信息失败: %v", err)
	}
	
	log.Printf("当前待上传文件数: %d, 总大小: %d bytes", 
		stats["total_files"], stats["total_size"])
	
	// TODO: 实现手动上传逻辑
	// 可以调用 uploader 的内部方法来触发上传
	
	return nil
}*/



