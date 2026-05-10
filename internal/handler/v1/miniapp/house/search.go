package house

import (
	"house-manager/pkg/errcode"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *HouseHandler) Search(c *gin.Context) {
	var req searchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(err))
		return
	}

	result, err := h.service.Search(c.Request.Context(), req.toServiceInput())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toSearchResponse(result))
}
