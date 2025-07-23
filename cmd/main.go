package main

import (
	"NewsEyeTracking/internal/api/routes"
	"NewsEyeTracking/internal/database"
	"NewsEyeTracking/internal/service"
	"NewsEyeTracking/pkg/logger"
	"context"
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
	services.News.Stop() // 停止新闻服务的后台任务
	h.Stop()              // 停止 handlers 的统一缓存管理任务
	log.Println("缓存数据已保存")

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务器强制关闭: %v", err)
	}

	log.Println("服务器已退出")
}
