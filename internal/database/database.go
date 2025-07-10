package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sql.DB, error){
	db, err := sql.Open("postgres", databaseURL)

	if err != nil {
		return nil, fmt.Errorf("数据库连接失败，Connect 函数中: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库 ping 失败，说明没有正确打开: %v", err)
	}

	log.Println("数据库连接成功")
	return db, nil
}