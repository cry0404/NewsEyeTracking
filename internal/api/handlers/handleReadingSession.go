package handlers

import (
	"NewsEyeTracking/internal/models"
	"NewsEyeTracking/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EndReadingSession 结束阅读会话
// POST /api/v1/sessions/:session_id/end
func (h *Handlers) EndReadingSession(c *gin.Context) {
	// 从JWT中间件获取用户ID, 现在默认应该都会存在了，code 与 user强绑定
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
	
	// 获取会话ID，这里在处理之前应该在缓存中查询有没有，缓存中查询
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"会话ID不能为空",
			"缺少必要参数",
		))
		return
	}


	// 解析请求体
	var req models.EndSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"请求参数不正确",
			err.Error(),
		))
		return
	}
	//sessionID := req.SessionID.String()

	// 如果没有提供结束时间，使用当前时间
	if req.EndTime.IsZero() {
		req.EndTime = time.Now()
	}

	// 为写操作创建带超时的 context
	ctx, cancel := utils.WithWriteTimeout(c.Request.Context())
	defer cancel()

	// 如果请求中包含最后一批追踪数据，先保存到缓存
	if req.Data != nil && !req.Data.IsEmpty() {
		// 构造一个临时的SessionDataRequest来复用缓存逻辑
		// 使用 sessionID 从路径参数中获取，并转换为 UUID
		sessionUUID, err := uuid.Parse(sessionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrorCodeInvalidRequest,
				"会话ID格式不正确",
				err.Error(),
			))
			return
		}
		
		tempReq := models.SessionDataRequest{
			SessionID: &sessionUUID,
			Data:      req.Data,
			Timestamp: req.EndTime,
		}
		
		if err := h.addToTrackingCache(userID, &tempReq); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrorCodeInternalError,
				"保存最后一批追踪数据失败",
				err.Error(),
			))
			return
		}
	}

	// 调用服务层结束会话
	err := h.services.Session.EndReadingSession(ctx, sessionID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"结束会话失败",
			err.Error(),
		))
		return
	}

	// 返回成功响应
	response := gin.H{
		"message":    "会话已成功结束",
		"session_id": sessionID,
		"end_time":   req.EndTime,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}


// POST /api/v1/sessions/:session_id/data
func (h *Handlers) ProcessSessionData(c *gin.Context) {
	// 从JWT中间件获取用户ID
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

	// 获取会话ID
	sessionIDStr := c.Param("session_id")
	if sessionIDStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"会话ID不能为空",
			"缺少必要参数",
		))
		return
	}

	// 验证会话ID格式
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"会话ID格式不正确",
			err.Error(),
		))
		return
	}

	// 解析请求体
	var req models.SessionDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"请求参数不正确",
			err.Error(),
		))
		return
	}

	// 如果请求中没有时间戳，使用当前时间
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	// 验证请求数据
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"请求数据验证失败",
			err.Error(),
		))
		return
	}

	// 处理心跳包
	if req.IsHeartbeat() {
		// 返回心跳包响应
		response := models.NewHeartbeatResponse()
		c.JSON(http.StatusOK, models.SuccessResponse(response))
		return
	}

	// 处理眼动数据
	if req.HasTrackingData() {
		// 设置会话ID（从URL参数）
		req.SessionID = &sessionID

		// 添加到缓存
		if err := h.addToTrackingCache(userID, &req); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrorCodeInternalError,
				"保存追踪数据失败",
				err.Error(),
			))
			return
		}

		// 返回数据处理成功响应
		response := models.NewDataResponse(sessionID)
		c.JSON(http.StatusOK, models.SuccessResponse(response))
		return
	}

	// 如果到达这里，说明请求无效
	c.JSON(http.StatusBadRequest, models.ErrorResponse(
		models.ErrorCodeInvalidRequest,
		"无效的请求数据",
		"请求必须是心跳包或包含有效的追踪数据",
	))
}

