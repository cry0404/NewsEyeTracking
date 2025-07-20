package handlers

import (
	"NewsEyeTracking/internal/models"
	"NewsEyeTracking/internal/service"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Handlers 处理程序结构体，持有所有依赖项
type Handlers struct {
	services      *service.Services
	trackingCache map[string][]models.UserTrackingRecord // 用户ID -> 追踪记录列表
	newsCache     map[string][]models.UserNewsRecord     // 用户ID -> 新闻记录列表
	cacheMutex    sync.RWMutex
	lastFlush     time.Time
	flushTicker   *time.Ticker
}

// 一个架构设计，通过 handler 来处理所有定义的服务
func NewHandlers(services *service.Services) *Handlers {
	h := &Handlers{
		services:      services,
		trackingCache: make(map[string][]models.UserTrackingRecord),
		newsCache:     make(map[string][]models.UserNewsRecord),
		lastFlush:     time.Now(),
		flushTicker:   time.NewTicker(30 * time.Second), // 每30秒刷新一次
	}
	
	// 启动后台统一缓存刷新任务
	go h.flushCacheRoutine()
	
	return h
}


func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}


func (h *Handlers) Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": "1.0.0"})
}

// flushCacheRoutine 后台定期刷新缓存数据
func (h *Handlers) flushCacheRoutine() {
	for range h.flushTicker.C {
		h.flushTrackingCache()
		h.flushNewsCache()
	}
}

// addToTrackingCache 将追踪数据添加到内存缓存
func (h *Handlers) addToTrackingCache(userID string, req *models.SessionDataRequest) error {
	if req == nil || req.SessionID == nil || req.Data == nil {
		return fmt.Errorf("追踪数据不完整")
	}
	
	h.cacheMutex.Lock()
	defer h.cacheMutex.Unlock()
	
	// 创建追踪记录
	record := models.UserTrackingRecord{
		SessionID: *req.SessionID,
		StartTime: req.Timestamp,
		Data:      *req.Data,
	}
	
	
	h.trackingCache[userID] = append(h.trackingCache[userID], record)
	
	return nil
}

// flushTrackingCache 将追踪缓存数据写入文件
func (h *Handlers) flushTrackingCache() {
	h.cacheMutex.Lock()
	defer h.cacheMutex.Unlock()
	
	if len(h.trackingCache) == 0 {
		return
	}
	
	// 创建今天的日期目录 - 统一所有追踪数据（眼动、点击、滚动）
	today := time.Now().Format("2006-01-02")
	dateDir := fmt.Sprintf("data/tracking/%s", today)
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		fmt.Printf("警告: 无法创建追踪数据目录: %v\n", err)
		return
	}
	
	// 为每个用户批量写入数据
	for userID, records := range h.trackingCache {
		if len(records) == 0 {
			continue
		}
		
		// 创建批量记录结构
		batchRecord := models.UserTrackingBatchRecord{
			FlushTime: time.Now(),
			Records:   records,
		}
		
		// 文件名使用用户ID
		fileName := fmt.Sprintf("%s.json", userID)
		filePath := fmt.Sprintf("%s/%s", dateDir, fileName)
		
		// 如果文件已存在，则追加到现有记录中
		if err := h.appendToTrackingFile(filePath, batchRecord); err != nil {
			fmt.Printf("警告: 无法写入用户%s的追踪数据文件: %v\n", userID, err)
			continue
		}
		
		// 统计事件数量
		totalEyeEvents := 0
		totalClickEvents := 0
		totalScrollEvents := 0
		for _, record := range records {
			totalEyeEvents += len(record.Data.EyeEvents)
			totalClickEvents += len(record.Data.ClickEvents)
			totalScrollEvents += len(record.Data.ScrollEvents)
		}
		
		fmt.Printf("成功写入用户%s的%d条追踪记录（眼动:%d, 点击:%d, 滚动:%d）\n", 
			userID, len(records), totalEyeEvents, totalClickEvents, totalScrollEvents)
	}
	
	// 通过新创建来手动更新缓存
	h.trackingCache = make(map[string][]models.UserTrackingRecord)
	h.lastFlush = time.Now()
}

// appendToTrackingFile 将新的批量记录追加到追踪数据文件中
func (h *Handlers) appendToTrackingFile(filePath string, newBatchRecord models.UserTrackingBatchRecord) error {
	var existingData models.UserTrackingBatchRecord
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err == nil {
		// 文件存在，读取现有数据
		existingBytes, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("无法读取现有文件: %w", err)
		}
		
		// 解析现有数据
		if err := json.Unmarshal(existingBytes, &existingData); err != nil {
			return fmt.Errorf("无法解析现有文件: %w", err)
		}
		
		// 合并记录
		existingData.Records = append(existingData.Records, newBatchRecord.Records...)
		existingData.FlushTime = newBatchRecord.FlushTime // 更新最后刷新时间
	} else {
		// 文件不存在，使用新数据
		existingData = newBatchRecord
	}
	
	// 序列化合并后的数据
	jsonData, err := json.MarshalIndent(existingData, "", "  ")
	if err != nil {
		return fmt.Errorf("无法序列化合并后的数据: %w", err)
	}
	
	// 写入文件
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("无法写入文件: %w", err)
	}
	
	return nil
}

// flushNewsCache 将新闻缓存数据写入文件
func (h *Handlers) flushNewsCache() {
	h.cacheMutex.Lock()
	defer h.cacheMutex.Unlock()
	
	if len(h.newsCache) == 0 {
		return
	}
	
	// 创建今天的日期目录
	today := time.Now().Format("2006-01-02")
	dateDir := fmt.Sprintf("data/news/%s", today)
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		fmt.Printf("警告: 无法创建新闻数据目录: %v\n", err)
		return
	}
	
	// 为每个用户批量写入数据
	for userID, records := range h.newsCache {
		if len(records) == 0 {
			continue
		}
		
		// 创建批量记录结构
		batchRecord := models.UserNewsBatchRecord{
			FlushTime: time.Now(),
			Records:   records,
		}
		
		// 文件名使用用户ID
		fileName := fmt.Sprintf("%s.json", userID)
		filePath := fmt.Sprintf("%s/%s", dateDir, fileName)
		
		// 如果文件已存在，则追加到现有记录中
		if err := h.appendToNewsFile(filePath, batchRecord); err != nil {
			fmt.Printf("警告: 无法写入用户%s的新闻数据文件: %v\n", userID, err)
			continue
		}
		
		fmt.Printf("成功写入用户%s的%d条新闻记录\n", userID, len(records))
	}
	
	// 清空缓存
	h.newsCache = make(map[string][]models.UserNewsRecord)
	h.lastFlush = time.Now()
}

// appendToNewsFile 将新的批量记录追加到新闻数据文件中
func (h *Handlers) appendToNewsFile(filePath string, newBatchRecord models.UserNewsBatchRecord) error {
	var existingData models.UserNewsBatchRecord
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err == nil {
		// 文件存在，读取现有数据
		existingBytes, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("无法读取现有文件: %w", err)
		}
		
		// 解析现有数据
		if err := json.Unmarshal(existingBytes, &existingData); err != nil {
			return fmt.Errorf("无法解析现有文件: %w", err)
		}
		
		// 合并记录
		existingData.Records = append(existingData.Records, newBatchRecord.Records...)
		existingData.FlushTime = newBatchRecord.FlushTime // 更新最后刷新时间
	} else {
		// 文件不存在，使用新数据
		existingData = newBatchRecord
	}
	
	// 序列化合并后的数据
	jsonData, err := json.MarshalIndent(existingData, "", "  ")
	if err != nil {
		return fmt.Errorf("无法序列化合并后的数据: %w", err)
	}
	
	// 写入文件
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("无法写入文件: %w", err)
	}
	
	return nil
}

// AddToNewsCache 将新闻数据添加到内存缓存
func (h *Handlers) AddToNewsCache(userID string, newsGUIDs []string) {
	h.cacheMutex.Lock()
	defer h.cacheMutex.Unlock()
	
	record := models.UserNewsRecord{
		StartTime: time.Now(),
		NewsGUIDs: newsGUIDs,
	}
	
	h.newsCache[userID] = append(h.newsCache[userID], record)
}

// Stop 停止后台任务并刷新缓存
func (h *Handlers) Stop() {
	h.flushTicker.Stop()
	h.FlushCaches() // 最后一次刷新
}


func (h *Handlers) FlushCaches() {
	h.flushTrackingCache()
	h.flushNewsCache()
}
