package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"NewsEyeTracking/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiterOptions 配置限流器的参数
type RateLimiterOptions struct {
	// 每秒允许的请求数
	RequestsPerSecond float64
	// 突发请求上限
	Burst int
	// 如何区分请求方的键：默认优先使用 userID，否则使用客户端 IP
	KeyResolver func(c *gin.Context) string
	// 跳过限流的路径（前缀匹配），例如健康检查
	SkipPaths []string
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiterStore struct {
	mu              sync.Mutex
	clients         map[string]*clientLimiter
	limit           rate.Limit
	burst           int
	cleanupInterval time.Duration
	entryTTL        time.Duration
}

func newRateLimiterStore(rps float64, burst int) *rateLimiterStore {
	s := &rateLimiterStore{
		clients:         make(map[string]*clientLimiter),
		limit:           rate.Limit(rps),
		burst:           burst,
		cleanupInterval: time.Minute,
		entryTTL:        5 * time.Minute,
	}
	go s.cleanupLoop()
	return s
}

func (s *rateLimiterStore) getLimiter(key string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cl, ok := s.clients[key]; ok {
		cl.lastSeen = time.Now()
		return cl.limiter
	}
	l := rate.NewLimiter(s.limit, s.burst)
	s.clients[key] = &clientLimiter{limiter: l, lastSeen: time.Now()}
	return l
}

func (s *rateLimiterStore) cleanupLoop() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-s.entryTTL)
		s.mu.Lock()
		for k, v := range s.clients {
			if v.lastSeen.Before(cutoff) {
				delete(s.clients, k)
			}
		}
		s.mu.Unlock()
	}
}

// RateLimit 基于可配置参数的通用限流中间件
func RateLimit(opts RateLimiterOptions) gin.HandlerFunc {
	// 默认配置
	if opts.RequestsPerSecond <= 0 {
		opts.RequestsPerSecond = 10
	}
	if opts.Burst <= 0 {
		opts.Burst = 20
	}
	if opts.KeyResolver == nil {
		opts.KeyResolver = func(c *gin.Context) string {
			if userID, ok := c.Get("userID"); ok {
				if s, ok2 := userID.(string); ok2 && s != "" {
					return "user:" + s
				}
			}
			return "ip:" + c.ClientIP()
		}
	}

	store := newRateLimiterStore(opts.RequestsPerSecond, opts.Burst)

	shouldSkip := func(path string) bool {
		for _, p := range opts.SkipPaths {
			if path == p || strings.HasPrefix(path, p) {
				return true
			}
		}
		return false
	}

	return func(c *gin.Context) {
		// 放行跳过路径
		if shouldSkip(c.Request.URL.Path) {
			c.Next()
			return
		}

		key := opts.KeyResolver(c)
		limiter := store.getLimiter(key)
		if !limiter.Allow() {
			// 告知客户端限流
			c.Header("Retry-After", "1")
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse(
				models.ErrorCodeTooManyRequests,
				"请求过于频繁",
				"请稍后再试",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitDefault 提供一个开箱即用的默认限流中间件
func RateLimitDefault() gin.HandlerFunc {
	return RateLimit(RateLimiterOptions{
		RequestsPerSecond: 10,
		Burst:             20,
		SkipPaths:         []string{"/api/v1/health", "/api/v1/version"},
	})
}
