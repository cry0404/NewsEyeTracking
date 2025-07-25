package handlers

import (
	"NewsEyeTracking/internal/models"
	"NewsEyeTracking/internal/utils"
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetNews 获取新闻列表
// GET /api/v1/news
func (h *Handlers) GetNews(c *gin.Context) {
	// 从JWT中间件获取用户ID， 这里的用户 id 还是字符串，需要解析成 uuid
	//暂时先注释掉， 还没有实现 jwt 之前

	// 这里似乎得写个中间件获取请求头中的分辨率？
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
	defaultLimit := req.Limit
	// 默认限制为5条，最多10条
	if req.Limit == 0 {
		req.Limit = 5
	} else{
		if req.Limit > 10 {
			req.Limit = 10
		}else{
			req.Limit = defaultLimit
		}
	}

	// 测试 id  ，实际应该填写对应的 userid， 先硬编码上去再说
	// 为数据库查询创建带超时的 context， 请求返回之前应该创建一个新的用户会话
	ctx, cancel := utils.WithDatabaseTimeout(c.Request.Context())
	defer cancel()
	// 这里是返回的新闻，也许 gjc 那边返回的就是 guid 呢
	newsList, err := h.services.News.GetNews(ctx, userID, req.Limit, h.AddToNewsCache)
	//统计 guid 并保存,直接记录保存也行？
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"获取新闻列表失败",
			"请先检查请求格式",
		))
		return
	}
	//能成功返回新闻列表了，就可以加入会话请求了
	/*用户行为记录
	每次刷新出来的新闻 ID 和顺序，当前屏幕分辨率
	用户在列表页的眼动浏览信息，当浏览到新闻标题和简述时要求给出分词实时数据，浏览其他内容需要给出对应的组件数据。所有数据需要标注是否开启了注视反馈。
	用户的点击数据，包含点击了哪条新闻，以及点击换一批新闻按钮的时机。

	在获取到内容后可以返回之前先创建一个阅读会话
	*/
	
	
	// 构建会话请求
	
	sessionReq := &models.CreateSessionRequestForArticles{
		StartTime:  time.Now(),
		DeviceInfo: utils.ExtractDeviceInfoFromHeaders(c), // 从请求中获取设备信息
	}
	// 创建列表页会话
	sessionForList, err := h.services.Session.CreateSessionForList(ctx, userID, sessionReq)
	
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"阅读会话创建失败",
			err.Error(),
		))
		return
	}
	
	// 构建响应，包含会话信息
	response := map[string]interface{}{
		"news_list": newsList,
	}
	
	// 如果会话创建成功，添加会话信息到响应中
	if sessionForList != nil {
		response["session"] = sessionForList
	}
	
	//返回新闻列表和会话信息
	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// GetNewsDetail 获取新闻详情
// GET /api/v1/news/:id
func (h *Handlers) GetNewsDetail(c *gin.Context) {
	// 从JWT中间件获取用户ID（用于A/B测试判断）
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			models.ErrorCodeUnauthorized,
			"未找到用户信息",
			"用户未认证",
		))
		return
	}
	
	// 解析用户ID
	userID, ok := userIDRaw.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"用户ID类型错误",
			"无法解析用户ID",
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
	
	useruuid, _ := uuid.Parse(userID)

	newsDetail, err := h.services.News.GetNewsDetail(ctx, newsIDStr, useruuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"获取新闻详情失败",
			err.Error(),
		))
		return
	}

	// 创建新闻详情页会话
	sessionReq := &models.CreateSessionRequestForArticle{
		ArticleID:  newsIDStr, // 使用新闻的GUID作为文章ID
		StartTime:  time.Now(),
		DeviceInfo: utils.ExtractDeviceInfoFromHeaders(c),
	}

	// 创建详情页会话
	sessionForFeed, err := h.services.Session.CreateSessionForFeed(ctx, userID, sessionReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"阅读会话创建失败",
			"服务器内部发生错误，请通知运营人员",
		))
		return
	}

	response := map[string]interface{}{
		"news_detail": newsDetail,
	}

	// 如果会话创建成功，添加会话信息到响应中
	if sessionForFeed != nil {
		response["session"] = sessionForFeed
	}

	// 返回新闻详情和会话信息
	c.JSON(http.StatusOK, models.SuccessResponse(response))
}
