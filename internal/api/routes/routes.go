package routes

import (
	"NewsEyeTracking/internal/api/handlers"
	"NewsEyeTracking/internal/api/middleware"
	"NewsEyeTracking/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, services *service.Services) {
	//全局处都需要使用到的中间件就定义到 middleware

	router.Use(middleware.ErrorHandler())

	router.Use(middleware.CORS())
	
	router.Use(middleware.Logger())
	

	h := handlers.NewHandlers(services)
	
	v1 := router.Group("/api/v1")
	{

		public := v1.Group("")
		{

			public.GET("/version", h.Version)
			public.GET("/health", h.HealthCheck)
			// 用户注册（无需认证）
			//public.POST("/auth/register", h.Register)
			public.POST("/auth/codes/:code", h.ValidCode) // 没有经过中间件可以直接解析， 业务逻辑是每次都需要输入邀请码登录，发放一个新的 jwt
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
			
			// 用户会话管理（简化版）
			//protected.POST("/session/init", h.InitUserSession)           // 整体的启动，应该整合在 code 路由中去，而且 init 中就应该判断会话状态
			//protected.POST("/session/data", h.ProcessSessionData)       // 统一数据上传接口，但这里是文章页的逻辑来

			
			// 会话管理
			//protected.POST("/sessions", h.CreateSession)
			//protected.PATCH("/sessions/:session_id", h.EndSession)

			//压缩数据上报
			//protected.POST("/sessions/:session_id/data", h.UploadCompressedData)
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
}
