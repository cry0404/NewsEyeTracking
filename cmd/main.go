package main

import (
	"log"
	"dabase/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("项目环境变量未成功加载")
	}

	dbURL := os.Getenv("DB_URL")

	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defet db.Close()

	services := service.NewServices(db)

}