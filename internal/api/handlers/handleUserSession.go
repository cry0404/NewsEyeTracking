package handlers

import (
	"NewsEyeTracking/internal/models"
	"NewsEyeTracking/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

)
// ProcessSessionData 处理会话数据，这里收到的数据应该没有完整的格式
//看需不需要另外做解析， 上传逻辑放到最后
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




/* 已废弃的实现
// InitUserSession 初始化用户会话， 用户会话需要的是保活机制
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
}*/
