package routes

import (
	"NewsEyeTracking/internal/api/handlers"
	"NewsEyeTracking/internal/api/middleware"
	"NewsEyeTracking/internal/service"
	"context"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, services *service.Services) *handlers.Handlers {
	//全局处都需要使用到的中间件就定义到 middleware
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.ReleaseMode // 默认使用生产模式
	}
	gin.SetMode(ginMode)

	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimitDefault())

	if ginMode != gin.ReleaseMode {
		router.Use(middleware.Logger())
	}

	// router.Use(middleware.DevLogger())
	// router.Use(middleware.SessionLog())

	h := handlers.NewHandlers(services)

	v1 := router.Group("/api/v1")
	{

		public := v1.Group("")
		{
			public.GET("/version", h.Version)
			public.GET("/health", h.HealthCheck)
			// 登录限流：在线人数上限（仅作用于创建会话的登录接口）
			maxOnline := 0
			if v := os.Getenv("MAX_ONLINE_USERS"); v != "" {
				if i, err := strconv.Atoi(v); err == nil && i > 0 {
					maxOnline = i
				}
			}
			if maxOnline > 0 {
				public.POST("/auth/codes/:code",
					middleware.OnlineLimit(maxOnline, func(ctx context.Context) (int, error) {
						return services.UserSession.GetActiveOnlineCount(ctx)
					}),
					h.ValidCode,
				)
			} else {
				public.POST("/auth/codes/:code", h.ValidCode)
			}
			// 用户登录是就应该有一个新的会话，我是否应该将第一个会话 id 与后面的内容联系起来呢
		}
		//还需要 session 管理
		protected := v1.Group("")
		// 测试研究的时候先禁用 jwt
		protected.Use(middleware.JWTAuth()) //这里每一个页面都需要有对应的 jwt
		{
			//用户相关
			protected.GET("/auth/profile", h.GetProfile)
			protected.GET("/auth/profile/", h.GetProfile)
			protected.POST("/auth/profile", h.UpdateProfile)
			protected.POST("/auth/profile/", h.UpdateProfile)
			protected.POST("/users/end", h.EndUserSession)
			// 用户会话管理（简化版）
			//protected.POST("/session/init", h.InitUserSession)           // 整体的启动，应该整合在 code 路由中去，而且 init 中就应该判断会话状态
			//      // 统一数据上传接口，但这里是文章页的逻辑来
			//protected.POST("/sessions/:session_id", h.EndReadingSession)

			// 会话管理
			//protected.POST("/sessions", h.CreateSession)

			//下面这个来结束会话？

			protected.POST("/sessions/:session_id/data", h.ProcessSessionData)

			protected.POST("/sessions/:session_id/end", h.EndReadingSession)
			//新闻相关
			newsProtected := protected.Group("")
			//newsProtected.Use(middleware.newsValid()) //这里实现的思路是把对应的 guid 区分开，检测天数，以免看太过时的新闻

			{

				newsProtected.GET("/news/", h.GetNews)
				newsProtected.GET("/news", h.GetNews)
				newsProtected.GET("/news/:id", h.GetNewsDetail)
				newsProtected.GET("/news/:id/", h.GetNewsDetail)
			}

		}

	}

	return h
}
