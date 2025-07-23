package service

import (
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/models"
	"context"
	"fmt"



	"github.com/google/uuid"
)

// NewsService 新闻列表的设计，主要是获
type NewsService interface {
	// GetNews 获取新闻列表
	GetNews(ctx context.Context, userID string, limit int, addToCache func(userID string, newsGUIDs []string)) (*models.NewsListResponse, error)
	// GetNewsDetail 获取新闻详情
	GetNewsDetail(ctx context.Context, newsID string) (*models.NewsDetailResponse, error)
	// Stop 停止后台任务并刷新缓存（现在为空实现，保持兼容性）
	Stop()
}

// newsService 新闻服务实现
type newsService struct {
	queries         *db.Queries
	recommendClient *RecommendService // 推荐服务客户端
}

// NewNewsService 创建新闻服务实例
func NewNewsService(queries *db.Queries, recommendClient *RecommendService) NewsService {
	return &newsService{
		queries:         queries,
		recommendClient: recommendClient,
	}
}

// GetNews 实现获取新闻列表逻辑, 服务端只需要实现服务逻辑就够了，不需要想着认证之类的东西
func (s *newsService) GetNews(ctx context.Context, userID string, limit int, addToCache func(userID string, newsGUIDs []string)) (*models.NewsListResponse, error) {

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

/*
	var hasRecommend bool
	// var hasMoreInformation bool

	hasRecommend = abConfig.HasRecommend.Bool

	if hasRecommend {
		//这里就是推荐算法的逻辑，返回的应该是
	}*/
	// 获取新闻列表

	//这里应该需要
	RecommendResponse, err := s.recommendClient.GetRecommendations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}
	var  articleID  []string
	for _, recommendNews := range RecommendResponse.Recommendations {
		articleID = append(articleID,  recommendNews.NewsID)
	}

	articles, err := s.queries.GetArticlesByGUID(ctx, articleID)

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
	if addToCache != nil {
		addToCache(userID, newsGUIDs)
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



// Stop 停止后台任务（简化实现，保持兼容性）
func (s *newsService) Stop() {
	// 由于缓存管理已经移动到 handlers，这里就是一个空实现
}


