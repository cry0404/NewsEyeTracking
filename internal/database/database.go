package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// RedisClient Redis客户端封装
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient 创建Redis客户端
func NewRedisClient() (*RedisClient, error) {
	// 从环境变量读取Redis配置
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	
	password := os.Getenv("REDIS_PASSWORD")
	
	dbStr := os.Getenv("REDIS_DB")
	db := 0
	if dbStr != "" {
		var err error
		db, err = strconv.Atoi(dbStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_DB value: %v", err)
		}
	}
	
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})
	
	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}
	
	log.Printf("Connected to Redis at %s:%s", host, port)
	
	return &RedisClient{client: rdb}, nil
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// GetClient 获取原始Redis客户端
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Set 设置键值对并指定过期时间
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取键值
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Exists 检查键是否存在
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	return result > 0, err
}

// Delete 删除键
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Expire 设置键的过期时间
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// Keys 获取匹配模式的所有键
func (r *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}
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
