package models

import (
	"time"
)

// News 新闻文章模型（对应rss系统中的feed_items表）
type News struct {
	ID          int        `json:"id" db:"id"`
	FeedID      int        `json:"feed_id" db:"feed_id"`
	Title       string     `json:"title" db:"title"`
	Description *string    `json:"description" db:"description"`
	Content     *string    `json:"content" db:"content"`
	Link        string     `json:"link" db:"link"`
	GUID        string     `json:"guid" db:"guid"`
	Author      *string    `json:"author" db:"author"`
	PublishedAt *time.Time `json:"published_at" db:"published_at"`
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// NewsListItem 新闻列表项响应（用于/news接口）
type NewsListItem struct {
	ID          int        `json:"id"`
	GUID        string     `json:"guid"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Author      *string    `json:"author"`
	PublishedAt *time.Time `json:"published_at"`
}

// NewsListResponse 新闻列表响应
type NewsListResponse struct {
	Articles []NewsListItem `json:"articles"`
	Limit    int            `json:"limit"`
}

// AdditionalInfo 额外信息（用于A/B测试）
type AdditionalInfo struct {
	LikeCount    int `json:"like_count"`
	CommentCount int `json:"comment_count"`
	ShareCount   int `json:"share_count"`
	SaveCount    int `json:"save_count"`
}
// 对应端点为 /new/{guid}
// NewsDetailResponse 新闻详情响应（支持A/B测试）, 默认为空
type NewsDetailResponse struct {
//这里不需要设置 id 了
	GUID           string          `json:"guid"`
	Title          string          `json:"title"`
	Description    *string         `json:"description"`
	Content        string          `json:"content"` // 包含class_id的HTML内容
	Author         *string         `json:"author"`
	PublishedAt    *time.Time      `json:"published_at"`
	AdditionalInfo *AdditionalInfo `json:"additional_info,omitempty"` // 根据A/B测试决定是否包含
}

// NewsRequest 新闻列表请求参数
type NewsRequest struct {
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}
