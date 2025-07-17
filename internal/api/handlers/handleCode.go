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
	token, err := h.services.User.UpdateLoginState(ctx, userID)//更新 jwt 的
	if err != nil{
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"获取邀请码失败",
			err.Error(),
		))
		return
	}
	//返回之前开始检查注册会话

	err = h.services.UserSession.CheckSingleSessionLimit(ctx, userID)
	if err != nil{
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.ErrorCodeForbidden,
			"登录失败，请确保同一时间只有一个账号使用",
			err.Error(),
		))
		return
	}//检测完后开始创建会话

	//那这里应该返回一个 session id ， user_session，然后保证可以根据这个查找
	newUserSession, err:= h.services.UserSession.CreateUserSession(ctx, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"服务器内部错误，请稍后尝试登入",
			err.Error(),
		))
		return
	}
	// token 中自带 user id，不用返回 user id
	c.JSON(http.StatusOK, models.SuccessResponse(models.LoginResponse{
		SessionID: newUserSession.SessionID,
		StartTime: newUserSession.StartTime,
		Token: token,
	}))
	
}