package middleware

import (
	"fmt"
	"strings"

	"house-manager/pkg/errcode"
	"house-manager/pkg/requestlog"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

func InternalAuth(expectedToken string) gin.HandlerFunc {
	expectedToken = strings.TrimSpace(expectedToken)
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("Request-ID"))
		if requestID == "" {
			response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("Request-ID is required")))
			c.Abort()
			return
		}
		requestlog.SetRequestID(c, requestID)
		c.Header("Request-ID", requestID)

		if expectedToken == "" {
			response.Err(c, errcode.InternalError.WithError(fmt.Errorf("internal token is not configured")))
			c.Abort()
			return
		}
		token := strings.TrimSpace(c.GetHeader("Internal-Token"))
		if token == "" || token != expectedToken {
			response.Err(c, errcode.Unauthorized.WithError(fmt.Errorf("invalid internal token")))
			c.Abort()
			return
		}
		c.Next()
	}
}
