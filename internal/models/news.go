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
	ID          	int        `json:"id"`
	GUID        	string     `json:"guid"`
	Title       	string     `json:"title"`
	Content			string     `json:"content"`
	LikeCount   	int32      `json:"like_count,omitempty"`
	ShareCount  	int32      `json:"share_count,omitempty"`
	SaveCount   	int32      `json:"save_count,omitempty"`
	CommentCount 	int32	   `json:"comment_count,omitempty"`
}

// NewsListResponse 新闻列表响应
type NewsListResponse struct {
	Articles []NewsListItem `json:"articles"`
	Limit    int            `json:"limit"`
}

// AdditionalInfo 额外信息（用于A/B测试）
type AdditionalInfo struct {
	LikeCount    int32 		`json:"like_count"`
	ShareCount   int32 		`json:"share_count"`
	SaveCount    int32 		`json:"save_count"`
	Comments     []Comment  `json:"comments"`
}

type Comment struct {
	Content		string      `json:"content"`
	Like  		int32   		`json:"like"`
	Replies		[]*Comment 	`json:"replies"`
}
// 对应端点为 /new/{guid}
// NewsDetailResponse 新闻详情响应（支持A/B测试）, 默认为空
type NewsDetailResponse struct {
//这里不需要设置 id 了
	GUID           string          `json:"guid"`
	Title          string          `json:"title"`
	//Description    *string         `json:"description"`
	Content        string          `json:"content"` // 包含class_id的HTML内容
	//Author         *string         `json:"author"`
	//PublishedAt    *time.Time      `json:"published_at"`
	AdditionalInfo *AdditionalInfo `json:"additional_info,omitempty"` // 根据A/B测试决定是否包含
}

// NewsRequest 新闻列表请求参数， 请求参数的设置根据 form 来做绑定， 可以有不同的参数来做解析
type NewsRequest struct {
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

// UserNewsRecord 用户新闻浏览记录
type UserNewsRecord struct {
	StartTime time.Time `json:"start_time"`
	NewsGUIDs []string  `json:"news_guids"`
}

// UserNewsBatchRecord 用户新闻批量记录（用于文件存储）
type UserNewsBatchRecord struct {
	FlushTime time.Time         `json:"flush_time"`
	Records   []UserNewsRecord `json:"records"`
}
