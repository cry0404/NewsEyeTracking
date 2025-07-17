package handlers

import (
	"NewsEyeTracking/internal/models"
	"NewsEyeTracking/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// handler 永远是用来处理某个路由产生的问题的
// 或者说处理某些服务的工具, handler 层通过 service 层的东西来解决
//在登录时就初始一个用户会话
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
	codeInfo, err := h.services.Auth.ValidateInviteCode(ctx, code) 
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"获取邀请码失败",
			"邀请码不存在",
		))
		return
	}

	/*if !codeInfo.IsUsed {
		//在没有注册的情况下， 然后等着在 createUser 部分等着做验证和创建

		c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
			"valid": "true",
		}))
		return
	}*/

	//这里就可以通过 userid 来返回类似于注册时的用户信息了, 无论如何第一次都应该返回，然后统一接口
	userID := codeInfo.ID
	token, err := h.services.User.UpdateLoginState(ctx, userID)
	if err != nil{
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"获取邀请码失败",
			err.Error(),
		))
		return
	}
	//返回之前开始注册会话
/*
	session, err := h.services.UserSession.CreateOrGetUserSession(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"创建或获取用户会话失败",
			err.Error(),
		))
		return
	}
*/	
	
	
	c.JSON(http.StatusOK, models.SuccessResponse(models.UserRegisterResponse{
		UserID: userID,
		Token: token,
	}))
	
}