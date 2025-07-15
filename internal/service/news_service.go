package service

import (
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/models"
	"context"
	"database/sql"
	"fmt"

	"time"

	"github.com/google/uuid"
)

// NewsService 新闻列表的设计，主要是获
type NewsService interface {
	// GetNews 获取新闻列表
	GetNews(ctx context.Context, userID string, limit int) (*models.NewsListResponse, error)
	// GetNewsDetail 获取新闻详情
	GetNewsDetail(ctx context.Context, newsID string) (*models.NewsDetailResponse, error)
}

// newsService 新闻服务实现
type newsService struct {
	queries *db.Queries
}

// NewNewsService 创建新闻服务实例
func NewNewsService(queries *db.Queries) NewsService {
	return &newsService{queries: queries}
}

// GetNews 实现获取新闻列表逻辑, 服务端只需要实现服务逻辑就够了，不需要想着认证之类的东西
func (s *newsService) GetNews(ctx context.Context, userID string, limit int) (*models.NewsListResponse, error) {
	// 1. 将 userID 转换为 UUID
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
	// 暂时直接获取最新的新闻
	articles, err := s.queries.GetNewArticles(ctx, db.GetNewArticlesParams{
		PublishedAt: sql.NullTime{Time: time.Now().AddDate(0, 0, -7), Valid: true}, // 最近7天的新闻
		Limit:       int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}

	// 4. 转换为响应格式
	newsItems := make([]models.NewsListItem, 0, len(articles))
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
	}

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
