package utils

import (
	"archive/zip"
	"context"
	"fmt"
	"io"

	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
)

// Config 上传器配置
type Config struct {
	WatchDir      string        // 监控目录
	UploadDir     string        // 临时上传目录
	MaxFiles      int           // 触发上传的最大文件数
	MaxSize       int64         // 触发上传的最大总大小(字节)
	CheckInterval time.Duration // 检查间隔
}

// FileUploader 文件上传器
type FileUploader struct {
	config    Config
	ossClient *oss.Client
	watcher   *fsnotify.Watcher
}

// NewUploader 创建新的上传器
func NewUploader(config Config) *FileUploader {
	return &FileUploader{
		config: config,
	}
}

// Start 启动上传器
func (u *FileUploader) Start(ctx context.Context) error {
	// 初始化OSS客户端
	if err := u.initOSSClient(); err != nil {
		return fmt.Errorf("初始化OSS客户端失败: %v", err)
	}

	// 创建文件监控器
	var err error
	u.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("创建文件监控器失败: %v", err)
	}
	defer u.watcher.Close()

	// 添加监控目录
	if err := u.watcher.Add(u.config.WatchDir); err != nil {
		return fmt.Errorf("添加监控目录失败: %v", err)
	}

	// 启动定时检查
	ticker := time.NewTicker(u.config.CheckInterval)
	defer ticker.Stop()

	log.Printf("开始监控目录: %s", u.config.WatchDir)

	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-u.watcher.Events:
			u.handleFileEvent(event)
		case err := <-u.watcher.Errors:
			log.Printf("文件监控错误: %v", err)
		case <-ticker.C:
			u.checkAndUpload()
		}
	}
}

// initOSSClient 初始化OSS客户端
func (u *FileUploader) initOSSClient() error {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，使用系统环境变量")
	}

	accessKeyId := os.Getenv("ACCESS_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY")
	if accessKeyId == "" || accessKeySecret == "" {
		return fmt.Errorf("请设置 ACCESS_ID 和 ACCESS_KEY 环境变量")
	}

	region := "cn-shenzhen" 
	provider := credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret)
	cfg := oss.LoadDefaultConfig().WithCredentialsProvider(provider).WithRegion(region)
	u.ossClient = oss.NewClient(cfg)

	return nil
}

// handleFileEvent 处理文件事件
func (u *FileUploader) handleFileEvent(event fsnotify.Event) {
	// 这里可以处理文件创建、修改等事件
	// 目前只做日志记录
	log.Printf("文件事件: %s %s", event.Op, event.Name)
}

// checkAndUpload 检查并上传文件
func (u *FileUploader) checkAndUpload() {
	files, totalSize, err := u.scanWatchDir()
	if err != nil {
		log.Printf("扫描监控目录失败: %v", err)
		return
	}

	// 检查是否达到上传条件
	if len(files) >= u.config.MaxFiles || totalSize >= u.config.MaxSize {
		log.Printf("达到上传条件 - 文件数: %d, 总大小: %d bytes", len(files), totalSize)
		if err := u.uploadFiles(files); err != nil {
			log.Printf("上传文件失败: %v", err)
		}
	}
}

// scanWatchDir 扫描监控目录
func (u *FileUploader) scanWatchDir() ([]string, int64, error) {
	var files []string
	var totalSize int64

	err := filepath.Walk(u.config.WatchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理普通文件，跳过目录
		if info.IsDir() {
			return nil
		}

		// 可以在这里添加文件过滤逻辑
		// 例如：只处理特定扩展名的文件
		files = append(files, path)
		totalSize += info.Size()
		return nil
	})

	return files, totalSize, err
}

// uploadFiles 上传文件
func (u *FileUploader) uploadFiles(files []string) error {
	if len(files) == 0 {
		return nil
	}

	// 创建压缩文件
	timestamp := time.Now().Format("20060102_150405")
	zipFileName := fmt.Sprintf("eyetracking_batch_%s.zip", timestamp)
	zipPath := filepath.Join(u.config.UploadDir, zipFileName)

	if err := u.createZipFile(files, zipPath); err != nil {
		return fmt.Errorf("创建压缩文件失败: %v", err)
	}

	// 上传到OSS
	if err := u.uploadToOSS(zipPath, zipFileName); err != nil {
		return fmt.Errorf("上传到OSS失败: %v", err)
	}


	if err := u.cleanupFiles(files); err != nil {
		log.Printf("清理文件失败: %v", err)
	}


	if err := os.Remove(zipPath); err != nil {
		log.Printf("删除临时压缩文件失败: %v", err)
	}

	log.Printf("成功上传 %d 个文件", len(files))
	return nil
}

// createZipFile 创建压缩文件
func (u *FileUploader) createZipFile(files []string, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		if err := u.addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}

	return nil
}

// addFileToZip 将文件添加到压缩包
func (u *FileUploader) addFileToZip(zipWriter *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 获取文件信息
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// 创建zip文件头
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// 设置文件名（使用相对路径）
	relPath, err := filepath.Rel(u.config.WatchDir, filename)
	if err != nil {
		relPath = filepath.Base(filename)
	}
	header.Name = relPath


	header.Method = zip.Deflate

	// 创建写入器
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// 复制文件内容
	_, err = io.Copy(writer, file)
	return err
}

// uploadToOSS 上传到阿里云OSS
func (u *FileUploader) uploadToOSS(filePath, objectName string) error {
	bucketName := "newseyetrackingtest" // 可以从环境变量读取

	putRequest := &oss.PutObjectRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectName),
		ProgressFn: func(increment, transferred, total int64) {
			progress := float64(transferred) / float64(total) * 100
			log.Printf("上传进度: %.2f%%", progress)
		},
	}

	result, err := u.ossClient.PutObjectFromFile(context.TODO(), putRequest, filePath)
	if err != nil {
		return err
	}

	log.Printf("文件上传成功: %s -> %s", filePath, *result.ETag)
	return nil
}

// cleanupFiles 清理已处理的文件
func (u *FileUploader) cleanupFiles(files []string) error {
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			log.Printf("删除文件失败 %s: %v", file, err)
		}
	}
	return nil
}

// GetStats 获取统计信息
func (u *FileUploader) GetStats() (map[string]interface{}, error) {
	files, totalSize, err := u.scanWatchDir()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_files":    len(files),
		"total_size":     totalSize,
		"watch_dir":      u.config.WatchDir,
		"max_files":      u.config.MaxFiles,
		"max_size":       u.config.MaxSize,
		"check_interval": u.config.CheckInterval.String(),
	}

	return stats, nil
}