package response

import (
	"net/http"

	"house-manager/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
	Data  any    `json:"data"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: 0, Data: data})
}

func Err(c *gin.Context, err error) {
	e := errcode.FromError(err)
	if e == nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 50001, Error: "服务内部错误"})
		return
	}
	c.JSON(toHTTPStatus(e.Code), Response{Code: e.Code, Error: e.Message})
}

func SuccessPage(c *gin.Context, data any, size int, maxSize int64) {
	c.JSON(http.StatusOK, PageResponse{Code: 0, MaxSize: maxSize, Size: size, Data: data})
}

type PageResponse struct {
	Code    int    `json:"code"`
	Error   string `json:"error"`
	MaxSize int64  `json:"maxSize"`
	Size    int    `json:"size"`
	Data    any    `json:"data"`
}

func toHTTPStatus(code int) int {
	switch code {
	case 10001:
		return http.StatusBadRequest
	case 10002:
		return http.StatusUnauthorized
	case 10003:
		return http.StatusForbidden
	case 10004:
		return http.StatusTooManyRequests
	case 10005:
		return http.StatusNotFound
	case 10006:
		return http.StatusConflict
	default:
		if code >= 50000 {
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}
}
