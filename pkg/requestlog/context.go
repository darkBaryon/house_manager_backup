package requestlog

import (
	"context"

	"github.com/gin-gonic/gin"
)

const (
	ContextRequestID    = "request_id"
	ContextResponseCode = "response_code"
	ContextResponseErr  = "response_error"
	ContextErrorDetail  = "response_error_detail"
	ContextFields       = "request_log_fields"
)

type requestIDContextKey struct{}

func SetRequestID(c *gin.Context, requestID string) {
	if c == nil || requestID == "" {
		return
	}
	c.Set(ContextRequestID, requestID)
	c.Request = c.Request.WithContext(ContextWithRequestID(c.Request.Context(), requestID))
}

func RequestID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	value, ok := c.Get(ContextRequestID)
	if !ok {
		return ""
	}
	requestID, _ := value.(string)
	return requestID
}

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil || requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDContextKey{}, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func SetResponse(c *gin.Context, code int, message string, detail string) {
	if c == nil {
		return
	}
	c.Set(ContextResponseCode, code)
	c.Set(ContextResponseErr, message)
	if detail != "" {
		c.Set(ContextErrorDetail, detail)
	}
}

func ResponseCode(c *gin.Context) int {
	if c == nil {
		return 0
	}
	value, ok := c.Get(ContextResponseCode)
	if !ok {
		return 0
	}
	code, _ := value.(int)
	return code
}

func ResponseError(c *gin.Context) string {
	if c == nil {
		return ""
	}
	value, ok := c.Get(ContextResponseErr)
	if !ok {
		return ""
	}
	message, _ := value.(string)
	return message
}

func ErrorDetail(c *gin.Context) string {
	if c == nil {
		return ""
	}
	value, ok := c.Get(ContextErrorDetail)
	if !ok {
		return ""
	}
	detail, _ := value.(string)
	return detail
}

func AddField(c *gin.Context, key string, value any) {
	if c == nil || key == "" || value == nil {
		return
	}
	fields := Fields(c)
	fields[key] = value
	c.Set(ContextFields, fields)
}

func AddFields(c *gin.Context, values map[string]any) {
	if c == nil || len(values) == 0 {
		return
	}
	fields := Fields(c)
	for key, value := range values {
		if key == "" || value == nil {
			continue
		}
		fields[key] = value
	}
	c.Set(ContextFields, fields)
}

func Fields(c *gin.Context) map[string]any {
	if c == nil {
		return map[string]any{}
	}
	value, ok := c.Get(ContextFields)
	if ok {
		if fields, ok := value.(map[string]any); ok && fields != nil {
			return fields
		}
	}
	return map[string]any{}
}
