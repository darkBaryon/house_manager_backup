package publish

import (
	"log/slog"

	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

func writePublishResult(c *gin.Context, logMessage string, data any, err error) {
	if err != nil {
		slog.Error(logMessage, "error", err)
		response.Err(c, err)
		return
	}
	response.Success(c, data)
}
