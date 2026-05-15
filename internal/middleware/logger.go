package middleware

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"house-manager/pkg/requestlog"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gin-gonic/gin"
)

// Logger 结构化请求日志中间件（基于 slog）
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		requestID := requestIDFromHeader(c.GetHeader("X-Request-ID"))
		requestlog.SetRequestID(c, requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		appCode := requestlog.ResponseCode(c)
		appError := requestlog.ResponseError(c)
		errorDetail := requestlog.ErrorDetail(c)
		displayPath := c.FullPath()
		if displayPath == "" {
			displayPath = path
		}
		message := fmt.Sprintf("%s %s %d %dms", c.Request.Method, displayPath, status, latency.Milliseconds())
		attrs := []any{
			"request_id", requestID,
		}
		if principal, ok := principalFromContext(c); ok {
			attrs = append(attrs,
				"principal", fmt.Sprintf("%s:%s:%s", principal.Terminal, principal.PrincipalType, maskPhone(principal.Phone)),
			)
		}

		switch {
		case status >= 500 || appCode >= 50000:
			slog.Error(message, append(attrs, diagnosticAttrs(c, path, query, appCode, appError, errorDetail)...)...)
		case status >= 400 || appCode > 0:
			slog.Warn(message, append(attrs, diagnosticAttrs(c, path, query, appCode, appError, errorDetail)...)...)
		default:
			slog.Info(message, attrs...)
		}
	}
}

func diagnosticAttrs(c *gin.Context, path, query string, appCode int, appError, errorDetail string) []any {
	attrs := []any{
		"path", path,
		"handler", c.HandlerName(),
		"app_code", appCode,
		"client_ip", c.ClientIP(),
	}
	if query != "" {
		attrs = append(attrs, "query", query)
	}
	if appError != "" {
		attrs = append(attrs, "app_error", appError)
	}
	if errorDetail != "" {
		attrs = append(attrs, "error_detail", errorDetail)
	}
	for key, value := range requestlog.Fields(c) {
		if key == "action" {
			attrs = append(attrs, "action", value)
			continue
		}
		attrs = append(attrs, "req_"+key, value)
	}
	return attrs
}

func principalFromContext(c *gin.Context) (session.Principal, bool) {
	value, ok := c.Get(ContextPrincipal)
	if !ok {
		return session.Principal{}, false
	}
	principal, ok := value.(session.Principal)
	return principal, ok
}

func requestIDFromHeader(value string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return bson.NewObjectID().Hex()
}

func maskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
