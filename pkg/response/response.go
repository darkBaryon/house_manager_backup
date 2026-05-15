package response

import (
	"net/http"

	"house-manager/pkg/errcode"
	"house-manager/pkg/requestlog"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
	Data  any    `json:"data"`
}

func Success(c *gin.Context, data any) {
	requestlog.SetResponse(c, 0, "", "")
	c.JSON(http.StatusOK, Response{Code: 0, Data: data})
}

func Err(c *gin.Context, err error) {
	e := errcode.FromError(err)
	if e == nil {
		requestlog.SetResponse(c, 50001, "服务内部错误", errString(err))
		c.JSON(http.StatusInternalServerError, Response{Code: 50001, Error: "服务内部错误"})
		return
	}
	requestlog.SetResponse(c, e.Code, e.Message, errString(err))
	c.JSON(toHTTPStatus(e.Code), Response{Code: e.Code, Error: e.Message})
}

func SuccessPage(c *gin.Context, data any, size int, maxSize int64) {
	requestlog.SetResponse(c, 0, "", "")
	c.JSON(http.StatusOK, PageResponse{Code: 0, MaxSize: maxSize, Size: size, Data: data})
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
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
