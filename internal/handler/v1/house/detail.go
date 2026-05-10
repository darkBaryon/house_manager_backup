package house

import (
	housesvc "house-manager/internal/service/house"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *HouseHandler) PublicDetail(c *gin.Context) {
	listingID, ok := bindDetailRequest(c)
	if !ok {
		return
	}

	result, err := h.service.GetPublicDetail(c.Request.Context(), housesvc.DetailInput{ListingID: listingID})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toDetailResponse(result))
}
