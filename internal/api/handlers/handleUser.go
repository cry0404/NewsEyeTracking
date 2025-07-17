package handlers

import (
	"NewsEyeTracking/internal/models"
	"NewsEyeTracking/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Register 处理用户注册，现在不需要 register 方法了
// POST /api/v1/auth/register
/*
func (h *Handlers) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"请求参数不正确",
			err.Error(),
		))
		return
	}

	//只需要在 createuser 部分标记使用就可以了，在邀请码登录部分增加次数
	ctx, cancel := utils.WithAuthTimeout(c.Request.Context())
	defer cancel()
	
	registerUser, err := h.services.User.CreateUser(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"注册失败",
			err.Error(),
		))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusCreated, models.SuccessResponse(registerUser))
}*/

// GetProfile 获取用户个人资料
// GET /api/v1/auth/profile
func (h *Handlers) GetProfile(c *gin.Context) {
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

	// 调用服务层获取用户信息
	// 为数据库查询创建带超时的 context
	ctx, cancel := utils.WithDatabaseTimeout(c.Request.Context())
	defer cancel()
	
	user, err := h.services.User.GetUserByID(ctx, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"获取用户信息失败",
			"用户尚未填写过信息",
		))
		return
	}

	// 返回用户信息
	c.JSON(http.StatusOK, models.SuccessResponse(user))
}

// UpdateProfile 更新用户个人资料
// POST /api/v1/auth/profile ， 现在统一两个接口
func (h *Handlers) UpdateProfile(c *gin.Context) {
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

	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"请求参数不正确",
			err.Error(),
		))
		return
	}

	// 调用服务层更新用户信息
	// 为写操作创建带超时的 context
	ctx, cancel := utils.WithWriteTimeout(c.Request.Context())
	defer cancel()
	
	user, err := h.services.User.UpdateUser(ctx, userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"更新用户信息失败",
			err.Error(),
		))
		return
	}

	// 返回更新后的用户信息
	c.JSON(http.StatusOK, models.SuccessResponse(user))
}


