package middleware

import (
	"strings"

	"house-manager/pkg/errcode"
	"house-manager/pkg/response"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
)

// Auth Redis Session 认证中间件，从 Authorization: Bearer <token> 提取并验证
func Auth(store *session.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			response.Err(c, errcode.Unauthorized)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		if tokenStr == auth {
			response.Err(c, errcode.Unauthorized)
			c.Abort()
			return
		}

		userId, err := store.Get(c.Request.Context(), tokenStr)
		if err != nil || userId == "" {
			response.Err(c, errcode.Unauthorized)
			c.Abort()
			return
		}

		c.Set("userId", userId)
		c.Next()
	}
}
