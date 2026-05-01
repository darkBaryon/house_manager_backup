package middleware

import (
	"log/slog"

	"house-manager/pkg/errcode"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

// Recovery panic 恢复中间件，返回统一 JSON 响应
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					"error", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)
				response.Err(c, errcode.InternalError)
				c.Abort()
			}
		}()
		c.Next()
	}
}
