package handlers

import (
	
)
/*
//将这个部分逻辑合并到 news 当中去
func (h *Handlers) CreateSession(c *gin.Context) {
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

	var req models.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"请求参数不正确",
			err.Error(),
		))
		return
	}

	// 调用服务层创建会话
	// 为写操作创建带超时的 context
	ctx, cancel := utils.WithWriteTimeout(c.Request.Context())
	defer cancel()
	
	session, err := h.services.Session.CreateSession(ctx, userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"创建会话失败",
			err.Error(),
		))
		return
	}

	// 返回会话信息
	c.JSON(http.StatusCreated, models.SuccessResponse(session))
}

// EndSession 这个部分应该考虑
func (h *Handlers) EndSession(c *gin.Context) {
	// 从JWT中间件获取用户ID（验证权限）
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			models.ErrorCodeUnauthorized,
			"未找到用户信息",
			"用户未认证",
		))
		return
	}

	// 获取会话ID
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"会话ID不能为空",
			"缺少必要参数",
		))
		return
	}

	var req models.EndSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"请求参数不正确",
			err.Error(),
		))
		return
	}

	// 调用服务层结束会话
	// 为写操作创建带超时的 context
	ctx, cancel := utils.WithWriteTimeout(c.Request.Context())
	defer cancel()
	
	err := h.services.Session.EndSession(ctx, sessionID, &req)
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
*/