package handlers

import (
	"NewsEyeTracking/internal/models"
	"NewsEyeTracking/internal/utils"
	"net/http"


	"github.com/gin-gonic/gin"
)

// GetNews 获取新闻列表
// GET /api/v1/news
func (h *Handlers) GetNews(c *gin.Context) {
	// 从JWT中间件获取用户ID， 这里的用户 id 还是字符串，需要解析成 uuid
	//暂时先注释掉， 还没有实现 jwt 之前
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			models.ErrorCodeUnauthorized,
			"未找到用户信息",
			"用户未认证",
		))
		return
	}

	
	userID, ok := userIDRaw.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"用户ID类型错误",
			"无法解析用户ID",
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
	if req.Limit == 0 || req.Limit > 20{
		req.Limit = 5
	}

	// 测试 id  ，实际应该填写对应的 userid， 先硬编码上去再说
	// 为数据库查询创建带超时的 context， 请求返回之前应该创建一个新的用户会话
	ctx, cancel := utils.WithDatabaseTimeout(c.Request.Context())
	defer cancel()
	
	newsList, err := h.services.News.GetNews(ctx, userID, req.Limit)
	//统计 guid 并保存,直接记录保存也行？
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"获取新闻列表失败",
			err.Error(),
		))
		return
	}
	//能成功返回新闻列表了，就可以加入会话请求了
	//var sessionreq models.CreateSessionRequest
	/*用户行为记录
	每次刷新出来的新闻 ID 和顺序，当前屏幕分辨率
	用户在列表页的眼动浏览信息，当浏览到新闻标题和简述时要求给出分词实时数据，浏览其他内容需要给出对应的组件数据。所有数据需要标注是否开启了注视反馈。
	用户的点击数据，包含点击了哪条新闻，以及点击换一批新闻按钮的时机。
*/


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

	// 验证ID格式，这里应该是验证 uuid 才对，到时候回过头来修改
	/*if _, err := strconv.Atoi(newsIDStr); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"新闻ID格式不正确",
			err.Error(),
		))
		return
	}*/

	// 调用服务层获取新闻详情
	// 为数据库查询创建带超时的 context
	ctx, cancel := utils.WithDatabaseTimeout(c.Request.Context())
	defer cancel()
	
	newsDetail, err := h.services.News.GetNewsDetail(ctx, newsIDStr)
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
