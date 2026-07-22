package middleware

import (
	"fmt"
	"slices"
	"strings"

	"house-manager/pkg/errcode"
	"house-manager/pkg/requestlog"
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

// AdminAuth 只允许后台管理 session。
func AdminAuth(store *session.Store) gin.HandlerFunc {
	return PrincipalAuth(store, session.TerminalAdmin)
}

// RequirePermission 校验当前登录 principal 是否包含指定权限码。
func RequirePermission(permissionCode string, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissionCode = strings.TrimSpace(permissionCode)
		if permissionCode == "" {
			response.Err(c, errcode.Forbidden.WithError(fmt.Errorf("当前账号无权执行该操作")))
			c.Abort()
			return
		}
		if strings.TrimSpace(message) == "" {
			message = "当前账号无权执行该操作"
		}

		value, ok := c.Get(ContextPrincipal)
		if !ok {
			response.Err(c, errcode.Unauthorized.WithError(fmt.Errorf("未登录或登录已过期，请重新登录")))
			c.Abort()
			return
		}
		principal, ok := value.(session.Principal)
		if !ok {
			response.Err(c, errcode.Unauthorized.WithError(fmt.Errorf("当前登录状态无效，请重新登录")))
			c.Abort()
			return
		}
		if !slices.Contains(principal.PermissionCodes, permissionCode) {
			response.Err(c, errcode.Forbidden.WithError(fmt.Errorf("%s", message)))
			c.Abort()
			return
		}
		c.Next()
	}
}

// PrincipalAuth 读取 Redis session principal，并按 terminal 做边界校验。
func PrincipalAuth(store *session.Store, terminal string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, ok := bearerToken(c)
		if !ok {
			setAuthFailure(c, "missing_bearer", terminal, nil, nil)
			response.Err(c, errcode.Unauthorized.WithError(fmt.Errorf("未登录或登录已过期，请重新登录")))
			c.Abort()
			return
		}

		principal, err := store.GetPrincipal(c.Request.Context(), tokenStr)
		if err != nil {
			setAuthFailure(c, "session_check_failed", terminal, nil, err)
			response.Err(c, errcode.CacheError.WithError(err))
			c.Abort()
			return
		}
		if principal == nil {
			setAuthFailure(c, "session_not_found", terminal, nil, nil)
			response.Err(c, errcode.Unauthorized.WithError(fmt.Errorf("未登录或登录已过期，请重新登录")))
			c.Abort()
			return
		}
		if terminal != "" && principal.Terminal != terminal {
			setAuthFailure(c, "terminal_mismatch", terminal, principal, nil)
			response.Err(c, errcode.Unauthorized.WithError(fmt.Errorf("当前登录状态无效，请重新登录")))
			c.Abort()
			return
		}

		setPrincipal(c, tokenStr, *principal)
		c.Next()
	}
}

func setAuthFailure(c *gin.Context, reason string, expectedTerminal string, principal *session.Principal, err error) {
	fields := map[string]any{
		"auth_failure": reason,
	}
	if expectedTerminal != "" {
		fields["auth_terminal_expected"] = expectedTerminal
	}
	if principal != nil {
		fields["auth_terminal_actual"] = principal.Terminal
		fields["auth_principal_type"] = principal.PrincipalType
		fields["auth_principal_id"] = principal.PrincipalID
	}
	if err != nil {
		fields["auth_error"] = err.Error()
	}
	requestlog.AddFields(c, fields)
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
