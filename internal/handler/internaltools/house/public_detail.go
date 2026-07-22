package house

import (
	housesvc "house-manager/internal/service/miniapp/house"
	"house-manager/pkg/errcode"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) PublicDetail(c *gin.Context) {
	var req publicDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(err))
		return
	}
	listingID, err := req.validate()
	if err != nil {
		response.Err(c, err)
		return
	}
	result, err := h.service.GetPublicDetail(c.Request.Context(), housesvc.DetailInput{ListingID: listingID})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toPublicDetailResponse(result))
}
