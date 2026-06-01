package middleware

import (
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
		if requestID == "" {
			requestID = requestIDFromHeader(c.GetHeader("Request-ID"))
		}
		if requestID == "" {
			requestID = bson.NewObjectID().Hex()
		}
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
		attrs := []any{
			"request_id", requestID,
			"method", c.Request.Method,
			"path", displayPath,
			"status", status,
			"duration", latency,
		}
		if appCode != 0 {
			attrs = append(attrs, "code", appCode)
		}
		if principal, ok := principalFromContext(c); ok {
			attrs = append(attrs,
				"principal", compactPrincipal(principal),
			)
		}

		switch {
		case status >= 500 || appCode >= 50000:
			slog.Error("http.request", append(attrs, diagnosticAttrs(c, path, query, appError, errorDetail)...)...)
		case status >= 400 || appCode > 0:
			slog.Warn("http.request", append(attrs, diagnosticAttrs(c, path, query, appError, errorDetail)...)...)
		default:
			slog.Info("http.request", attrs...)
		}
	}
}

func diagnosticAttrs(c *gin.Context, path, query string, appError, errorDetail string) []any {
	attrs := []any{
		"client_ip", c.ClientIP(),
		"raw_path", path,
		"handler", c.HandlerName(),
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
	return ""
}

func maskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

func compactPrincipal(principal session.Principal) string {
	parts := []string{
		strings.TrimSpace(principal.Terminal),
		strings.TrimSpace(principal.PrincipalType),
	}
	if phone := maskPhone(principal.Phone); phone != "" {
		parts = append(parts, phone)
	} else if principal.PrincipalID != "" {
		parts = append(parts, principal.PrincipalID)
	}
	return strings.Join(parts, ":")
}
