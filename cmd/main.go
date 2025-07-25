package main

import (
	"NewsEyeTracking/internal/api/routes"
	"NewsEyeTracking/internal/database"
	"NewsEyeTracking/internal/service"
	"NewsEyeTracking/pkg/logger"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// 初始化环境变量
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("项目环境变量未成功加载，将使用系统环境变量")
	}

	// 初始化zap日志记录器
	if err := logger.InitLogger(); err != nil {
		log.Fatal("初始化日志记录器失败:", err)
	}
	// 确保在程序退出时刷新日志缓冲区
	defer logger.Sync()

	// 获取数据库连接字符串
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		logger.Logger.Fatal("DB_URL环境变量未设置")
	}

	// 验证上传服务所需的环境变量
	if err := validateUploadEnvVars(); err != nil {
		logger.Logger.Fatal("上传服务环境变量验证失败", zap.Error(err))
	}

	// 创建上传服务所需的目录
	if err := createUploadDirectories(); err != nil {
		logger.Logger.Fatal("创建上传目录失败", zap.Error(err))
	}

	// 连接数据库
	db, err := database.Connect(dbURL)
	if err != nil {
		logger.Logger.Fatal("数据库连接失败", zap.Error(err))
	}
	defer db.Close()

	// 连接Redis
	redisClient, err := database.NewRedisClient()
	if err != nil {
		logger.Logger.Fatal("Redis连接失败", zap.Error(err))
	}
	defer redisClient.Close()

	services := service.NewServices(db, redisClient)

	// 创建上传服务的上下文
	uploadCtx, uploadCancel := context.WithCancel(context.Background())
	defer uploadCancel()

	// 启动上传服务监控
	if err := services.Upload.StartMonitoring(uploadCtx); err != nil {
		logger.Logger.Error("启动上传服务失败", zap.Error(err))
	} else {
		logger.Logger.Info("上传服务已启动")
	}

	r := gin.New()
	

	// 设置路由并获取 handlers 实例
	h := routes.SetupRoutes(r, services)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 创建一个goroutine来启动服务器
	go func() {
		log.Printf("服务器启动在端口 %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("启动服务器出错，请排查：%v", err)
		}
	}()

	// 等待中断信号来优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 优雅关闭服务器，等待当前请求完成
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭服务之前，先停止所有后台任务并刷新缓存
	log.Println("正在保存缓存数据...")
	uploadCancel()        // 停止上传服务
	services.News.Stop() // 停止新闻服务的后台任务
	h.Stop()              // 停止 handlers 的统一缓存管理任务
	log.Println("缓存数据已保存")

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务器强制关闭: %v", err)
	}

	log.Println("服务器已退出")
}

// validateUploadEnvVars 验证上传服务所需的环境变量
func validateUploadEnvVars() error {
	requiredVars := []string{
		"ACCESS_ID",
		"ACCESS_KEY",
		"OSS_REGION",
		"OSS_BUCKET_NAME",
		"UPLOAD_TRACKING_DIR",
		"UPLOAD_NEWS_DIR",
		"UPLOAD_TEMP_DIR",
		"UPLOAD_MAX_FILES",
		"UPLOAD_MAX_SIZE",
		"UPLOAD_CHECK_INTERVAL",
	}

	for _, varName := range requiredVars {
		if value := os.Getenv(varName); value == "" {
			return fmt.Errorf("环境变量 %s 未设置", varName)
		}
	}

	return nil
}

// createUploadDirectories 创建上传服务所需的目录
func createUploadDirectories() error {
	trackingDir := os.Getenv("UPLOAD_TRACKING_DIR")
	newsDir := os.Getenv("UPLOAD_NEWS_DIR")
	tempDir := os.Getenv("UPLOAD_TEMP_DIR")

	// 创建tracking目录
	if err := os.MkdirAll(trackingDir, 0755); err != nil {
		return fmt.Errorf("创建tracking目录 %s 失败: %v", trackingDir, err)
	}

	// 创建news目录
	if err := os.MkdirAll(newsDir, 0755); err != nil {
		return fmt.Errorf("创建news目录 %s 失败: %v", newsDir, err)
	}

	// 创建临时目录
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录 %s 失败: %v", tempDir, err)
	}

	log.Printf("上传目录创建成功 - tracking: %s, news: %s, 临时: %s", trackingDir, newsDir, tempDir)
	return nil
}
