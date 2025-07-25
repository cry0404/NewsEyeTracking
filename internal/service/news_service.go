package service

import (
	"NewsEyeTracking/internal/db"
	"NewsEyeTracking/internal/models"
	"context"
	//"database/sql"
	"fmt"
	"encoding/json"
	//"time"
	//"math/rand"


	"github.com/google/uuid"
)

// NewsService 新闻列表的设计，主要是获
type NewsService interface {
	// GetNews 获取新闻列表
	GetNews(ctx context.Context, userID string, limit int, addToCache func(userID string, newsGUIDs []string)) (*models.NewsListResponse, error)
	// GetNewsDetail 获取新闻详情
	GetNewsDetail(ctx context.Context, newsID string, userID uuid.UUID) (*models.NewsDetailResponse, error)
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


	// 安全处理HasMoreInformation字段
	hasMoreInformation := abConfig.HasMoreInformation.Valid && abConfig.HasMoreInformation.Bool
	//这里判断是否需要增加额外信息, 需要获取每一篇新闻的点赞收藏等信息

	 RecommendResponse, err := s.recommendClient.GetRecommendations(ctx, userID)
	 if err != nil {
	 	return nil, fmt.Errorf("failed to get articles: %w", err)
	 }
	 var  articleID  []string
	 for _, recommendNews := range RecommendResponse.Recommendations {
	 	articleID = append(articleID,  recommendNews.NewsID)
	 }
	 Params := db.GetArticlesByGUIDParams{
		Column1: articleID,
		Limit:   int32(limit),
	 }
	 articles, err := s.queries.GetArticlesByGUID(ctx, Params)

	// 临时解决方案：直接从数据库获取最新文章
	/*oneDayAgo := time.Now().AddDate(0, 0, -1)
	articles, err := s.queries.GetNewArticles(ctx, db.GetNewArticlesParams{
		PublishedAt: sql.NullTime{Time: oneDayAgo, Valid: true},
		Limit: int32(limit),
	})*/

	if err != nil {
		return nil, fmt.Errorf("从数据库查询对应文章错误: %w", err)
	}
	newsItems := make([]models.NewsListItem, 0, len(articles))
	newsGUIDs := make([]string, 0, len(articles))
	if hasMoreInformation {
		for _, article := range articles {
			newsItem := models.NewsListItem{
				ID:          int(article.ID),
				GUID:        article.Guid,
				Content:     article.Content.String,
				Title:       article.Title,
				LikeCount:   article.LikeCount.Int32,
				ShareCount:  article.ShareCount.Int32,
				SaveCount:   article.SaveCount.Int32,
				CommentCount:article.CommentCount.Int32 ,
			}
			
			

			newsItems = append(newsItems, newsItem)
			newsGUIDs = append(newsGUIDs, article.Guid)
		}
	}else{
		for _, article := range articles {
			newsItem := models.NewsListItem{
				ID:          int(article.ID),
				GUID:        article.Guid,
				Content:     article.Content.String,
				Title:       article.Title,
			}
			
			

			newsItems = append(newsItems, newsItem)
			newsGUIDs = append(newsGUIDs, article.Guid)
		}
	}


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
/*
type AdditionalInfo struct {
	LikeCount    int `json:"like_count"`
	CommentCount int `json:"comment_count"`
	ShareCount   int `json:"share_count"`
	SaveCount    int `json:"save_count"`
}*/

// GetNewsDetail 实现获取新闻详情逻辑
func (s *newsService) GetNewsDetail(ctx context.Context, newsID string, userID uuid.UUID) (*models.NewsDetailResponse, error) {
	//这里应该考虑是否添加点赞评论收藏
	//定义对应的数据结构
	abConfig, err := s.queries.GetABTestConfigByInviteCodeID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("暂时无法获取到内部的测试信息: %w", err)
	}


	// 安全处理HasMoreInformation字段
	hasMoreInformation := abConfig.HasMoreInformation.Valid && abConfig.HasMoreInformation.Bool
	//先默认调整为 false， 最后再来调整
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
	if hasMoreInformation {
		additionalInformation, err := s.queries.GetMoreInfoMation(ctx, article.Guid)
		if err != nil {
			return nil, fmt.Errorf("获取额外信息失败: %w", err)
		}
		
		// 初始化AdditionalInfo
		newsItem.AdditionalInfo = &models.AdditionalInfo{
			LikeCount:  additionalInformation.LikeCount.Int32,
			SaveCount:  additionalInformation.SaveCount.Int32,
			ShareCount: additionalInformation.ShareCount.Int32,
			Comments:   []models.Comment{}, // 默认为空数组
		}
		
		// 解析评论数据
		var comments []models.Comment
		if additionalInformation.Comments.Valid {
			err := json.Unmarshal(additionalInformation.Comments.RawMessage, &comments)
			if err != nil {
				return nil, fmt.Errorf("解析评论失败: %v", err)
			}
			newsItem.AdditionalInfo.Comments = comments
		}
	}
	
	return &newsItem, nil


}
//返回一个更多的评论
/*
func randomMoreInformation () (*models.AdditionalInfo, error){
//	var moreinformation  models.AdditionalInfo
	//定义一个 limit，需要匹配我要获取的 limit，给一个
	//randomInt := rand.Intn(1000)

	return nil, nil
}*/

// Stop 停止后台任务（简化实现，保持兼容性）
func (s *newsService) Stop() {
	


}


