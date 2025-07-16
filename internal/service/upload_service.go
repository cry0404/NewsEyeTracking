package service

import (
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/utils"
	"context"
	"fmt"
	"log"
	"time"
)

// UploadService 上传服务接口
type UploadService interface {
	// StartMonitoring 启动文件监控和自动上传
	StartMonitoring(ctx context.Context) error
	// UploadFiles 手动触发文件上传
	//UploadFiles(ctx context.Context) error
	// GetUploadStats 获取上传统计信息
	GetUploadStats(ctx context.Context) (map[string]interface{}, error)
	// BackupData 备份数据
	BackupData(ctx context.Context) error
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
	// 设置默认配置
	config := utils.Config{
		WatchDir:      "./data/eyetracking", // 监控的眼动数据目录
		UploadDir:     "./data/temp",        // 临时目录
		MaxFiles:      100,                  // 达到100个文件时触发上传
		MaxSize:       100 * 1024 * 1024,    // 达到100MB时触发上传
		CheckInterval: 5 * time.Minute,      // 每5分钟检查一次
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

// GetUploadStats 获取上传统计信息
func (s *uploadService) GetUploadStats(ctx context.Context) (map[string]interface{}, error) {
	// 从文件系统获取统计信息
	fileStats, err := s.uploader.GetStats()
	if err != nil {
		return nil, fmt.Errorf("获取文件统计信息失败: %v", err)
	}
	
	// TODO: 可以从数据库获取历史上传记录等信息
	// uploadHistory, err := s.queries.GetUploadHistory(ctx)
	
	// 合并统计信息
	stats := map[string]interface{}{
		"file_stats":    fileStats,
		"last_check":    time.Now().Format(time.RFC3339),
		"service_status": "running",
	}
	
	return stats, nil
}

// BackupData 备份数据
func (s *uploadService) BackupData(ctx context.Context) error {
	log.Println("开始数据备份...")
	
	// TODO: 实现数据备份逻辑
	// 1. 创建备份目录
	// 2. 复制重要数据
	// 3. 压缩备份文件
	// 4. 上传到OSS的备份bucket
	
	return nil
}

// CleanupOldData 清理过期数据
func (s *uploadService) CleanupOldData(ctx context.Context, daysToKeep int) error {
	log.Printf("清理 %d 天前的数据...", daysToKeep)
	
	// TODO: 实现数据清理逻辑
	// 1. 查询数据库中的过期记录
	// 2. 删除OSS上的对应文件
	// 3. 删除本地缓存文件
	// 4. 更新数据库记录
	
	return nil
}
