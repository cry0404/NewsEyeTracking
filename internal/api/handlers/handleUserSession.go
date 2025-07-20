package handlers

import (
	"NewsEyeTracking/internal/models"
	"NewsEyeTracking/internal/service"
	"net/http"

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

func (h *UserSessionHandler) HandleHeartbeat(c *gin.Context) {
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

// HandleGetSessionStatus 获取会话状态

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
}

// HandleEndSession 结束会话
func (h *UserSessionHandler) HandleEndSession(c *gin.Context) {
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

	// 调用服务层结束会话
	err = h.sessionService.EndUserSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrorCodeInternalError,
			"结束会话失败",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("会话已成功结束"))
}
