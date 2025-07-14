package main

import (
	"NewsEyeTracking/internal/api/routes"
	"NewsEyeTracking/internal/database"
	"NewsEyeTracking/internal/service"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("项目环境变量未成功加载，将使用系统环境变量")
	}

	// 获取数据库连接字符串
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL环境变量未设置")
	}

	// 连接数据库
	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 初始化服务层
	services := service.NewServices(db)

	// 创建Gin引擎
	r := gin.New()

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 设置路由
	routes.SetupRoutes(r, services)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("服务器启动在端口 %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Printf("启动服务器出错，请排查：%v", err)
	}
}
