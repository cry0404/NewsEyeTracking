package database

import (
	"database/sql"
	"fmt"
	"log"

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

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("数据库连接成功")
	return db, nil
}
