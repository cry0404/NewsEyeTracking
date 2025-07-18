package service

import (
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// NewsService 新闻列表的设计，主要是获
type NewsService interface {
	// GetNews 获取新闻列表
	GetNews(ctx context.Context, userID string, limit int) (*models.NewsListResponse, error)
	// GetNewsDetail 获取新闻详情
	GetNewsDetail(ctx context.Context, newsID string) (*models.NewsDetailResponse, error)
	// Stop 停止后台任务并刷新缓存
	Stop()
	// FlushCache 手动刷新缓存
	FlushCache()
}

// newsService 新闻服务实现
type newsService struct {
	queries     *db.Queries
	newsCache   map[string][]models.UserNewsRecord // 用户ID -> 记录列表
	cacheMutex  sync.RWMutex
	lastFlush   time.Time
	flushTicker *time.Ticker
}

// NewNewsService 创建新闻服务实例
func NewNewsService(queries *db.Queries) NewsService {
	service := &newsService{
		queries:     queries,
		newsCache:   make(map[string][]models.UserNewsRecord),
		lastFlush:   time.Now(),
		flushTicker: time.NewTicker(30 * time.Second), // 每30秒刷新一次
	}
	
	// 启动后台刷新任务， 只使用一个 go routine 来监督
	go service.flushCacheRoutine()
	
	return service
}

// GetNews 实现获取新闻列表逻辑, 服务端只需要实现服务逻辑就够了，不需要想着认证之类的东西
func (s *newsService) GetNews(ctx context.Context, userID string, limit int) (*models.NewsListResponse, error) {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("uuid 格式不正确: %w", err)
	}

	//  根据用户A/B测试配置决定是否应用推荐算法
	// userID 跟邀请码是一对一强绑定的, 暂时先不考率算法也行
	abConfig, err := s.queries.GetABTestConfigByInviteCodeID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("暂时无法获取到内部的测试信息: %w", err)
	}

	// 获取新闻列表
	// 如果启用推荐算法，这里可以调用推荐算法，在这里接入推荐算法？
articles, err := s.queries.GetRandomArticles(ctx, db.GetRandomArticlesParams{
		PublishedAt: sql.NullTime{Time: time.Now().AddDate(0, 0, -7), Valid: true}, // 最近7天的新闻
		Limit:       int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}


	newsItems := make([]models.NewsListItem, 0, len(articles))
	newsGUIDs := make([]string, 0, len(articles))
	for _, article := range articles {
		newsItem := models.NewsListItem{
			ID:          int(article.ID),
			GUID:        article.Guid,
			Title:       article.Title,
			Description: &article.Description.String,
			Author:      &article.Author.String,
			PublishedAt: &article.PublishedAt.Time,
		}

		// 处理可能为空的字段
		if !article.Description.Valid {
			newsItem.Description = nil
		}
		if !article.Author.Valid {
			newsItem.Author = nil
		}
		if !article.PublishedAt.Valid {
			newsItem.PublishedAt = nil
		}

		newsItems = append(newsItems, newsItem)
		newsGUIDs = append(newsGUIDs, article.Guid)
	}

	// 错误检查后，将用户浏览的新闻GUID添加到缓存
	s.addToCache(userID, newsGUIDs)

	// 根据A/B测试配置决定点赞收藏评论
	_ = abConfig // 暂时标记使用

	return &models.NewsListResponse{
		Articles: newsItems,
		Limit:    limit,
	}, nil
}

// GetNewsDetail 实现获取新闻详情逻辑
func (s *newsService) GetNewsDetail(ctx context.Context, newsID string) (*models.NewsDetailResponse, error) {

	article, err := s.queries.GetArticleByGUID(ctx, newsID)
	if err != nil {
        return nil, fmt.Errorf("failed to get article: %w", err)
    }
	newsItem := models.NewsDetailResponse{
		GUID:        article.Guid,
		Title:       article.Title,
		Description: nil, 
		Content:     "",  
		Author:      nil, 
		PublishedAt: nil, 
		
	}//这里的 id 默认字段为 0
	// 安全地处理可能为空的字段
	if article.Description.Valid {
		newsItem.Description = &article.Description.String
	}

	if article.Content.Valid {
		newsItem.Content = article.Content.String 
	}

	if article.Author.Valid {
		newsItem.Author = &article.Author.String
	}

	if article.PublishedAt.Valid {
		newsItem.PublishedAt = &article.PublishedAt.Time
	}

	return &newsItem, nil
}

// addToCache 将用户新闻记录添加到内存缓存, 可能内存泄露，还需要观察？
func (s *newsService) addToCache(userID string, newsGUIDs []string) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	record := models.UserNewsRecord{
		StartTime: time.Now(),
		NewsGUIDs: newsGUIDs,
	}
	
	s.newsCache[userID] = append(s.newsCache[userID], record)
}

// flushCacheRoutine 后台定期刷新缓存到文件
func (s *newsService) flushCacheRoutine() {
	for range s.flushTicker.C {
		s.flushCacheToFile()
	}
}

// flushCacheToFile 将缓存中的数据批量写入文件， 备选是直接写入，需要测试
// 这里的思路应该是可以将所有的前端发来的信息都存入缓存中，我现在需要处理的是心跳包会话问题
func (s *newsService) flushCacheToFile() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	if len(s.newsCache) == 0 {
		return
	}
	
	// 创建今天的日期目录
	today := time.Now().Format("2006-01-02")
	dateDir := fmt.Sprintf("data/user/%s", today)
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		fmt.Printf("警告: 无法创建日期目录: %v\n", err)
		return
	}
	
	// 为每个用户批量写入数据
	for userID, records := range s.newsCache {
		if len(records) == 0 {
			continue
		}
		
		// 创建批量记录结构
		batchRecord := models.UserNewsBatchRecord{
			FlushTime: time.Now(),
			Records:   records,
		}
		
		// 文件名仅使用用户ID
		fileName := fmt.Sprintf("%s.json", userID)
		filePath := fmt.Sprintf("%s/%s", dateDir, fileName)
		
		// 如果文件已存在，则追加到现有记录中
		if err := s.appendToUserFile(filePath, batchRecord); err != nil {
			fmt.Printf("警告: 无法写入用户%s的新闻记录文件: %v\n", userID, err)
			continue
		}
		
		fmt.Printf("成功写入用户%s的%d条新闻记录\n", userID, len(records))
	}
	
	// 清空缓存
	s.newsCache = make(map[string][]models.UserNewsRecord)
	s.lastFlush = time.Now()
}

// FlushCache 手动刷新缓存（可用于优雅关闭）
func (s *newsService) FlushCache() {
	s.flushCacheToFile()
}

// Stop 停止后台任务
func (s *newsService) Stop() {
	s.flushTicker.Stop()
	s.flushCacheToFile() // 最后一次刷新
}
// appendToUserFile 将新的批量记录追加到用户文件中
func (s *newsService) appendToUserFile(filePath string, newBatchRecord models.UserNewsBatchRecord) error {
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
