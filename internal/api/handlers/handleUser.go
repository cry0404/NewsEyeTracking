package handlers

import (
	"NewsEyeTracking/internal/models"
	"NewsEyeTracking/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Register 处理用户注册
// POST /api/v1/auth/register
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

	// 调用服务层处理注册逻辑，这里注册并返回一个 jwt
	// 为认证操作创建带超时的 context
	ctx, cancel := utils.WithAuthTimeout(c.Request.Context())
	defer cancel()
	
	user, err := h.services.User.CreateUser(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"注册失败",
			err.Error(),
		))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusCreated, models.SuccessResponse(user))
}

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
			err.Error(),
		))
		return
	}

	// 返回用户信息
	c.JSON(http.StatusOK, models.SuccessResponse(user))
}

// UpdateProfile 更新用户个人资料
// PUT /api/v1/auth/profile
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

	var req models.UserUpdateRequest
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


// handler 永远是用来处理某个路由产生的问题的
// 或者说处理某些服务的工具, handler 层通过 service 层的东西来解决
//POST /api/v1/auth/codes/{code}
func (h *Handlers) ValidCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"邀请码不能为空",
			"缺少必要参数",
		))
		return
	}
	ctx, cancel := utils.WithDatabaseTimeout(c.Request.Context())
	defer cancel()
	//查询数据库默认加一个查询 context 避免超时太多
	_, err := h.services.Auth.ValidateInviteCode(ctx, code) //返回来的不重要，只需要验证即可
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"获取邀请码失败",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"valid": "true",
	}))

}