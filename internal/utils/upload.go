package utils

//作为工具函数来使用， 这里主要考虑如何上传，是否需要上传到 oss 或者自建的图床？
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
	TrackingDir   string        // 眼动追踪数据目录
	NewsDir       string        // 新闻数据目录
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

	// 添加监控目录 - tracking 和 news
	if err := u.watcher.Add(u.config.TrackingDir); err != nil {
		return fmt.Errorf("添加tracking监控目录失败: %v", err)
	}
	if err := u.watcher.Add(u.config.NewsDir); err != nil {
		return fmt.Errorf("添加news监控目录失败: %v", err)
	}

	// 启动定时检查
	ticker := time.NewTicker(u.config.CheckInterval)
	defer ticker.Stop()

	log.Printf("开始监控目录: %s, %s", u.config.TrackingDir, u.config.NewsDir)

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

	region := os.Getenv("OSS_REGION")
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
	// 分别扫描 tracking 和 news 目录
	u.checkAndUploadDirectory(u.config.TrackingDir, "tracking")
	u.checkAndUploadDirectory(u.config.NewsDir, "news")
}

// checkAndUploadDirectory 检查并上传指定目录
func (u *FileUploader) checkAndUploadDirectory(dir, dirType string) {
	files, totalSize, err := u.scanDirectory(dir)
	if err != nil {
		log.Printf("扫描%s目录失败: %v", dirType, err)
		return
	}

	if len(files) == 0 {
		return
	}

	// 检查是否达到上传条件
	if len(files) >= u.config.MaxFiles || totalSize >= u.config.MaxSize {
		log.Printf("%s达到上传条件 - 文件数: %d, 总大小: %d bytes", dirType, len(files), totalSize)
		if err := u.uploadDirectoryFiles(files, dir, dirType); err != nil {
			log.Printf("上传%s文件失败: %v", dirType, err)
		}
	}
}

// scanDirectory 扫描指定目录
func (u *FileUploader) scanDirectory(dir string) ([]string, int64, error) {
	var files []string
	var totalSize int64

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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

// uploadDirectoryFiles 上传指定目录的文件
func (u *FileUploader) uploadDirectoryFiles(files []string, baseDir, dirType string) error {
	if len(files) == 0 {
		return nil
	}

	// 创建压缩文件
	timestamp := time.Now().Format("20060102_150405")
	zipFileName := fmt.Sprintf("%s_batch_%s.zip", dirType, timestamp)
	zipPath := filepath.Join(u.config.UploadDir, zipFileName)

	if err := u.createZipFileWithBaseDir(files, zipPath, baseDir); err != nil {
		return fmt.Errorf("创建压缩文件失败: %v", err)
	}

	// 上传到OSS的对应文件夹
	ossObjectName := fmt.Sprintf("%s/%s", dirType, zipFileName)
	if err := u.uploadToOSS(zipPath, ossObjectName); err != nil {
		return fmt.Errorf("上传到OSS失败: %v", err)
	}

	// 清理文件
	if err := u.cleanupFiles(files); err != nil {
		log.Printf("清理文件失败: %v", err)
	}

	// 删除临时压缩文件
	if err := os.Remove(zipPath); err != nil {
		log.Printf("删除临时压缩文件失败: %v", err)
	}

	log.Printf("成功上传 %s 类型 %d 个文件", dirType, len(files))
	return nil
}

// createZipFileWithBaseDir 创建带基础目录的压缩文件
func (u *FileUploader) createZipFileWithBaseDir(files []string, zipPath, baseDir string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		if err := u.addFileToZipWithBaseDir(zipWriter, file, baseDir); err != nil {
			return err
		}
	}

	return nil
}

// addFileToZipWithBaseDir 将文件添加到压缩包（保持目录结构）
func (u *FileUploader) addFileToZipWithBaseDir(zipWriter *zip.Writer, filename, baseDir string) error {
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

	// 设置文件名（使用相对路径，保持目录结构）
	relPath, err := filepath.Rel(baseDir, filename)
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
	bucketName := os.Getenv("OSS_BUCKET_NAME")

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

// ForceUploadAll 强制上传所有文件，不检查阈值条件
func (u *FileUploader) ForceUploadAll() error {
	log.Println("开始强制上传所有文件...")

	// 强制上传 tracking 目录
	if err := u.forceUploadDirectory(u.config.TrackingDir, "tracking"); err != nil {
		log.Printf("强制上传tracking目录失败: %v", err)
		return err
	}

	// 强制上传 news 目录
	if err := u.forceUploadDirectory(u.config.NewsDir, "news"); err != nil {
		log.Printf("强制上传news目录失败: %v", err)
		return err
	}

	log.Println("强制上传完成")
	return nil
}

// forceUploadDirectory 强制上传指定目录的所有文件
func (u *FileUploader) forceUploadDirectory(dir, dirType string) error {
	files, _, err := u.scanDirectory(dir)
	if err != nil {
		return fmt.Errorf("扫描%s目录失败: %v", dirType, err)
	}

	if len(files) == 0 {
		log.Printf("%s目录没有文件需要上传", dirType)
		return nil
	}

	log.Printf("强制上传%s目录 - 文件数: %d", dirType, len(files))
	return u.uploadDirectoryFiles(files, dir, dirType)
}

// GetStats 获取统计信息
func (u *FileUploader) GetStats() (map[string]interface{}, error) {
	// 分别获取tracking和news的统计信息
	trackingFiles, trackingSize, err := u.scanDirectory(u.config.TrackingDir)
	if err != nil {
		return nil, err
	}

	newsFiles, newsSize, err := u.scanDirectory(u.config.NewsDir)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"tracking_files": len(trackingFiles),
		"tracking_size":  trackingSize,
		"news_files":     len(newsFiles),
		"news_size":      newsSize,
		"total_files":    len(trackingFiles) + len(newsFiles),
		"total_size":     trackingSize + newsSize,
		"tracking_dir":   u.config.TrackingDir,
		"news_dir":       u.config.NewsDir,
		"max_files":      u.config.MaxFiles,
		"max_size":       u.config.MaxSize,
		"check_interval": u.config.CheckInterval.String(),
	}

	return stats, nil
}
