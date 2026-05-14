package middleware

import (
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
		tokenStr, ok := bearerToken(c)
		if !ok {
			c.Next()
			return
		}
		principal, err := store.GetPrincipal(c.Request.Context(), tokenStr)
		if err == nil && principal != nil && principal.Terminal == session.TerminalMiniapp && principal.PrincipalType == session.PrincipalTypeUser {
			setPrincipal(c, tokenStr, *principal)
		}
		c.Next()
	}
}
