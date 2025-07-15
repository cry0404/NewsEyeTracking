package database

import (
	"database/sql"
	"fmt"
	"log"
	"runtime"

	_ "github.com/lib/pq"
)

// Connect 连接到PostgreSQL数据库
func Connect(dbURL string) (*sql.DB, error) {
	if dbURL == "" {
		return nil, fmt.Errorf("数据库连接字符串不能为空")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 优化连接池参数以支持高并发
	maxOpenConns := runtime.NumCPU() * 10  // 动态计算
	if maxOpenConns < 50 { maxOpenConns = 50 }
	if maxOpenConns > 200 { maxOpenConns = 200 }

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxOpenConns / 4)

	log.Println("数据库连接成功")
	return db, nil
}
