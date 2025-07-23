package service

import (
	"NewsEyeTracking/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RecommendService 推荐服务客户端
type RecommendService struct {
	client  *http.Client
	baseURL string
}

// NewRecommendService 创建推荐服务客户端
func NewRecommendService(baseURL string) *RecommendService {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:6667" // Python Flask服务端口
	}

	return &RecommendService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

// GetRecommendations 获取用户推荐
func (r *RecommendService) GetRecommendations(ctx context.Context, userID string) (*models.RecommendResponse, error) {
	// 构建请求体，直接使用UUID字符串
	requestBody := map[string]interface{}{
		"user_id": userID,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		r.baseURL+"/recommend",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")


	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送推荐请求失败: %w", err)
	}
	defer resp.Body.Close()


	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取推荐响应失败: %w", err)
	}


	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("推荐服务返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}


	var recommendResp models.RecommendResponse
	if err := json.Unmarshal(body, &recommendResp); err != nil {
		return nil, fmt.Errorf("解析推荐响应失败: %w", err)
	}

	return &recommendResp, nil
}

// HealthCheck 检查推荐服务健康状态
func (r *RecommendService) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/", nil)
	if err != nil {
		return fmt.Errorf("创建健康检查请求失败: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("推荐服务健康检查失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("推荐服务健康检查失败，状态码: %d", resp.StatusCode)
	}

	return nil
}
