package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"NewsEyeTracking/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// MakeJWT 创建JWT令牌， 前端存储到 session storage 中去， 这样来管理会话
func MakeJWT(userID uuid.UUID, tokenSecret string, expireIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "Eyetracking",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),//每次登录都会颁发一个新的
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expireIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ValidateJWT 验证JWT令牌
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(tokenSecret), nil
		},
	)

	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid claims")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	return userID, nil
} // jwt 返回的也是一个 uuid

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从header中获取Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrorCodeUnauthorized,
				"缺少Authorization头",
				"请提供Bearer token",
			))
			c.Abort()
			return
		}

		// 检查Bearer前缀
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrorCodeUnauthorized,
				"无效的Authorization格式",
				"格式应为: Bearer <token>",
			))
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		err := godotenv.Load()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrorCodeInternalError,
				"服务器环境变量加载错误",
				err.Error(),
			))
		}
		jwtSecret := os.Getenv("JWT_SECRET") // 这里应该从配置文件或环境变量获取

		// 验证token
		userID, err := ValidateJWT(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrorCodeUnauthorized,
				"无效的token",
				err.Error(),
			))
			c.Abort()
			return
		}

		// 将用户ID存储到上下文中
		c.Set("userID", userID.String()) //方便根据需要提取，因为后面的验证需要
		c.Next()
	}
}
