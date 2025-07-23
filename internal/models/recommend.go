package models

// 推荐响应结构
type RecommendResponse struct {
	UserID          string               `json:"user_id"` // 支持UUID字符串
	Strategy        string               `json:"strategy"`
	Recommendations []RecommendationItem `json:"recommendations"`
}

// 推荐项结构
type RecommendationItem struct {
	NewsID string     `json:"news_id"`
	Score  float64 `json:"score"`
}

// 推荐请求结构
type RecommendRequest struct {
	UserID int `json:"user_id"`
}



