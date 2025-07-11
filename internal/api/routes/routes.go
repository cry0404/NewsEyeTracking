package routes

import (
	"NewsEyeTracking/internal/api/handlers"
	"NewsEyeTracking/internal/api/middleware"
	"NewsEyeTracking/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, services *service.Services) {
	//全局处都需要使用到的中间件就定义到 middleware

	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandler())

	h := handlers.NewHandlers(services)
	//按资源类型作为分类
	v1 := router.Group("/api/v1")
	{
		//按照公开可访问端点和私有端点做区分
		public := v1.Group("")
		{
			// 健康检查
			public.GET("/version", h.Version)
			public.GET("/health", h.HealthCheck)
			// 用户注册（无需认证）
			public.POST("/auth/register", h.Register)
		}
		//还需要 session 管理
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth()) //这里每一个页面都需要有对应的 jwt
		{
			//用户相关
			protected.GET("/auth/profile", h.GetProfile)
			protected.PUT("/auth/profile", h.UpdateProfile)

			//新闻相关
			protected.GET("/news", h.GetNews)
			protected.GET("/news/:id", h.GetNewsDetail)

			// 会话管理
			protected.POST("/sessions", h.CreateSession)
			protected.PATCH("/sessions/:session_id", h.EndSession)

			//压缩数据上报
			protected.POST("/sessions/:session_id/data", h.UploadCompressedData)
		}

	}
}
