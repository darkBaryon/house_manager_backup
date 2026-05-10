package middleware

import (
	"strings"

	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
)

// OptionalAuth 放行匿名请求；如果 Bearer token 有效，则写入 userId。
func OptionalAuth(store *session.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if store == nil {
			c.Next()
			return
		}
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.Next()
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		if tokenStr == auth || strings.TrimSpace(tokenStr) == "" {
			c.Next()
			return
		}
		userID, err := store.Get(c.Request.Context(), tokenStr)
		if err == nil && userID != "" {
			c.Set("userId", userID)
		}
		c.Next()
	}
}
