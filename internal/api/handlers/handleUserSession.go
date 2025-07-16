package handlers

import (
	"NewsEyeTracking/internal/models"
	"NewsEyeTracking/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// InitUserSession 初始化用户会话
// POST /api/v1/session/init
func (h *Handlers) InitUserSession(c *gin.Context) {
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

	// 为写操作创建带超时的 context
	ctx, cancel := utils.WithWriteTimeout(c.Request.Context())
	defer cancel()

	// 解析用户ID
	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"用户ID格式错误",
			err.Error(),
		))
		return
	}

	// 创建或获取用户会话
	session, err := h.services.UserSession.CreateOrGetUserSession(ctx, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"创建或获取用户会话失败",
			err.Error(),
		))
		return
	}

	// 返回会话信息
	c.JSON(http.StatusOK, models.SuccessResponse(session))
}

// ProcessSessionData 处理会话数据（统一处理心跳和眼动数据）
// POST /api/v1/session/data
func (h *Handlers) ProcessSessionData(c *gin.Context) {
	var req models.DataUploadRequest
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

	// 为文件操作创建带超时的 context
	ctx, cancel := utils.WithFileOperationTimeout(c.Request.Context())
	defer cancel()

	// 处理数据上传
	response, err := h.services.UserSession.ProcessDataUpload(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeDataProcessError,
			"数据上传处理失败",
			err.Error(),
		))
		return
	}

	// 返回处理结果
	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// GetCurrentSessionStatus 获取当前用户会话状态
// GET /api/v1/session/status
func (h *Handlers) GetCurrentSessionStatus(c *gin.Context) {
	// 从 JWT中间件获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			models.ErrorCodeUnauthorized,
			"未找到用户信息",
			"用户未认证",
		))
		return
	}

	// 解析用户ID
	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"用户ID格式错误",
			err.Error(),
		))
		return
	}

	// 为读操作创建带超时的 context
	ctx, cancel := utils.WithReadTimeout(c.Request.Context())
	defer cancel()

	// 获取或创建用户会话（这里不会真的创建，只会获取现有的）
	session, err := h.services.UserSession.CreateOrGetUserSession(ctx, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"获取用户会话失败",
			err.Error(),
		))
		return
	}

	// 获取会话状态
	status, err := h.services.UserSession.GetSessionStatus(ctx, session.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"获取会话状态失败",
			err.Error(),
		))
		return
	}

	// 返回会话状态
	c.JSON(http.StatusOK, models.SuccessResponse(status))
}

// EndUserSession 结束用户会话
// DELETE /api/v1/user-sessions/:session_id
func (h *Handlers) EndUserSession(c *gin.Context) {
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

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"会话ID格式错误",
			err.Error(),
		))
		return
	}

	// 为写操作创建带超时的 context
	ctx, cancel := utils.WithWriteTimeout(c.Request.Context())
	defer cancel()

	// 结束会话
	err = h.services.UserSession.EndUserSession(ctx, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"结束会话失败",
			err.Error(),
		))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"message": "会话已成功结束"}))
}
