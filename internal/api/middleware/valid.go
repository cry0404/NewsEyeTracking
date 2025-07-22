package middleware

import (
	"NewsEyeTracking/internal/models"
	"net/http"
	"time"


	"github.com/gin-gonic/gin"
)

//这里最后再做设计，有关核验这里的新闻是否合法，是否为最近几天的内容
func ValidNews() gin.HandlerFunc{
	return func(c *gin.Context){
		defer func(){
			newsid := c.Param("id")

			if newsid == "" {
				c.Next()
			}else{//处理一下不是当天新闻的日期
				today := time.Now().Format("20060102")

				requestDay := newsid[4:12]
				/*var is_equal bool
				if today == requestDay{
					is_equal = true
				}else{
					is_equal = false
				}
				log.Printf("当前的 today: %v, 当前的 requestDay: %v, 是否相等: %v", today, requestDay, is_equal)*/
				if today != requestDay {
					c.JSON(http.StatusBadRequest, models.ErrorResponse(
						models.ErrorBadRequest,
						"不提供访问 1 天前的新闻",
						"请提供正确的新闻访问参数",
					))
					c.Abort()
					return
				}
				c.Next()
			}
		}()
	}
}
