package middleware

import (
	"github.com/zxc7563598/fintrack-backend/jwt"
	"github.com/zxc7563598/fintrack-backend/utils/response"

	"github.com/gin-gonic/gin"
)

// 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			response.Fail(c, 300001)
			c.Abort()
			return
		}
		claims, err := jwt.ParseToken(tokenStr)
		if err != nil {
			response.Fail(c, 300002)
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
