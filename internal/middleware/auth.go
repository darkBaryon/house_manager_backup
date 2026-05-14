package middleware

import (
	"strings"

	"house-manager/pkg/errcode"
	"house-manager/pkg/response"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
)

const (
	ContextToken     = "token"
	ContextPrincipal = "principal"
	ContextUserID    = "userId"
)

// Auth Redis Session 认证中间件，从 Authorization: Bearer <token> 提取并验证
func Auth(store *session.Store) gin.HandlerFunc {
	return PrincipalAuth(store, "")
}

// MiniappAuth 只允许小程序 session。
func MiniappAuth(store *session.Store) gin.HandlerFunc {
	return PrincipalAuth(store, session.TerminalMiniapp)
}

// PublishAuth 只允许出房 Web session。
func PublishAuth(store *session.Store) gin.HandlerFunc {
	return PrincipalAuth(store, session.TerminalPublish)
}

// PrincipalAuth 读取 Redis session principal，并按 terminal 做边界校验。
func PrincipalAuth(store *session.Store, terminal string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, ok := bearerToken(c)
		if !ok {
			response.Err(c, errcode.Unauthorized)
			c.Abort()
			return
		}

		principal, err := store.GetPrincipal(c.Request.Context(), tokenStr)
		if err != nil || principal == nil {
			response.Err(c, errcode.Unauthorized)
			c.Abort()
			return
		}
		if terminal != "" && principal.Terminal != terminal {
			response.Err(c, errcode.Unauthorized)
			c.Abort()
			return
		}

		setPrincipal(c, tokenStr, *principal)
		c.Next()
	}
}

func setPrincipal(c *gin.Context, token string, principal session.Principal) {
	c.Set(ContextToken, token)
	c.Set(ContextPrincipal, principal)
	c.Request = c.Request.WithContext(session.ContextWithPrincipal(c.Request.Context(), principal))
	if principal.PrincipalType == session.PrincipalTypeUser {
		c.Set(ContextUserID, principal.PrincipalID)
	}
}

func bearerToken(c *gin.Context) (string, bool) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return "", false
	}
	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	if tokenStr == auth || strings.TrimSpace(tokenStr) == "" {
		return "", false
	}
	return tokenStr, true
}
