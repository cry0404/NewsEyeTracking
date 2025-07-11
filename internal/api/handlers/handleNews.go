package handlers

import (
	"NewsEyeTracking/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetNews 获取新闻列表
// GET /api/v1/news
func (h *Handlers) GetNews(c *gin.Context) {
	// 从JWT中间件获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			models.ErrorCodeUnauthorized,
			"未找到用户信息",
			"用户未认证",
		))
		return
	}

	// 获取查询参数
	var req models.NewsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"请求参数不正确",
			err.Error(),
		))
		return
	}

	// 默认限制为10条
	if req.Limit == 0 {
		req.Limit = 10
	}

	// 调用服务层获取新闻列表
	newsList, err := h.services.News.GetNews(c.Request.Context(), userID.(string), req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"获取新闻列表失败",
			err.Error(),
		))
		return
	}

	// 返回新闻列表
	c.JSON(http.StatusOK, models.SuccessResponse(newsList))
}

// GetNewsDetail 获取新闻详情
// GET /api/v1/news/:id
func (h *Handlers) GetNewsDetail(c *gin.Context) {
	// 从JWT中间件获取用户ID（用于A/B测试判断）
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			models.ErrorCodeUnauthorized,
			"未找到用户信息",
			"用户未认证",
		))
		return
	}

	// 获取新闻ID
	newsIDStr := c.Param("id")
	if newsIDStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"新闻ID不能为空",
			"缺少必要参数",
		))
		return
	}

	// 验证ID格式
	if _, err := strconv.Atoi(newsIDStr); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"新闻ID格式不正确",
			err.Error(),
		))
		return
	}

	// 调用服务层获取新闻详情
	newsDetail, err := h.services.News.GetNewsDetail(c.Request.Context(), newsIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"获取新闻详情失败",
			err.Error(),
		))
		return
	}

	// 返回新闻详情
	c.JSON(http.StatusOK, models.SuccessResponse(newsDetail))
}
