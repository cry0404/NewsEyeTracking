package handlers

import (
	"NewsEyeTracking/internal/models"
	"NewsEyeTracking/internal/service"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserSessionHandler struct {
	sessionService service.UserSessionService
}

func NewUserSessionHandler(sessionService service.UserSessionService) *UserSessionHandler {
	return &UserSessionHandler{
		sessionService: sessionService,
	}
}

// HandleHeartbeat 处理心跳请求

func (h *UserSessionHandler) HandleHeartbeat(c *gin.Context) {//在处理心跳包的时候应该还要调用这里 usersession 的内容来处理
	var req models.HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"无效的请求参数",
			err.Error(),
		))
		return
	}

	// 调用服务层处理心跳
	response, err := h.sessionService.Heartbeat(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"处理心跳请求失败",
			err.Error(),
		))
		return
	}

	// 根据心跳响应状态返回不同的HTTP状态码
	switch response.Status {
	case "ok":
		c.JSON(http.StatusOK, models.SuccessResponse(response))
	case "invalid":
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			models.ErrorCodeSessionNotFound,
			response.Message,
			"",
		))
	case "expired":
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			models.ErrorCodeUnauthorized,
			response.Message,
			"",
		))
	default:
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"未知的会话状态",
			"",
		))
	}
}

// EndUserSession 结束会话并登出用户
// POST /users/end - 这个接口同时处理会话结束和用户登出
func (h *Handlers) EndUserSession(c *gin.Context) {
	// 获取用户ID
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


	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"用户ID格式错误",
			err.Error(),
		))
		return
	}


	authHeader := c.GetHeader("Authorization")
	var jwtToken string
	if authHeader != "" {
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
			jwtToken = tokenParts[1]
		}
	}

	// 获取用户的活跃会话
	activeSession, err := h.services.UserSession.GetActiveUserSessionByUserID(c.Request.Context(), userUUID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// 即使没有活跃会话，也要处理JWT登出
			if jwtToken != "" {
				// 这里可以将JWT加入黑名单或进行其他登出处理
				//前端登出即可
			}
			c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
				"message": "用户已登出（无活跃会话）",
				"logout_required": true, // 提示前端清除JWT token
			}))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"获取会话信息失败",
			err.Error(),
		))
		return
	}

	// 结束该用户所有活跃的阅读会话
	activeSessions, err := h.services.Session.GetUserActiveSessions(c.Request.Context(), userUUID)
	if err != nil {
		// 即使获取阅读会话失败，也继续结束用户会话
		fmt.Printf("获取用户活跃阅读会话失败: %v\n", err)
	}

	// 结束所有活跃的阅读会话
	var endedReadingSessions []string
	for _, readingSession := range activeSessions {
		// 使用当前时间作为结束时间
		endReq := &models.EndSessionRequest{
			EndTime: time.Now(),
		}
		
		// 结束该阅读会话
		if endErr := h.services.Session.EndReadingSession(c.Request.Context(), readingSession.ID.String(), endReq); endErr != nil {
			// 记录错误但不阻断流程
			fmt.Printf("结束阅读会话 %s 失败: %v\n", readingSession.ID.String(), endErr)
		} else {
			endedReadingSessions = append(endedReadingSessions, readingSession.ID.String())
		}
	}

	// 结束用户会话
	err = h.services.UserSession.EndUserSession(c.Request.Context(), activeSession.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"结束用户会话失败",
			err.Error(),
		))
		return
	}

	// JWT登出处理
	// 对于简单的无状态JWT，我们采用以下策略：
	// 1. 告知前端立即清除JWT token
	// 2. 如果需要更严格的安全性，可以实现JWT黑名单机制
	
	// 返回成功响应，包含登出指示和结束的阅读会话信息
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "用户会话已成功结束，用户已登出",
		"user_session_id": activeSession.ID,
		"ended_reading_sessions": endedReadingSessions,
		"ended_reading_sessions_count": len(endedReadingSessions),
		"logout_required": true, // 提示前端清除JWT token
	}))
}



// HandleGetSessionStatus 获取会话状态
/*
func (h *UserSessionHandler) HandleGetSessionStatus(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrorCodeInvalidRequest,
			"无效的会话ID格式",
			err.Error(),
		))
		return
	}

	// 调用服务层获取会话状态
	status, err := h.sessionService.GetSessionStatus(c.Request.Context(), sessionID)
	if err != nil {
		if err.Error() == "会话不存在" {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrorCodeSessionNotFound,
				"会话不存在",
				"",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"获取会话状态失败",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(status))
}*/