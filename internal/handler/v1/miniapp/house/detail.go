package house

import (
	minihandler "house-manager/internal/handler/v1/miniapp/common"
	housesvc "house-manager/internal/service/miniapp/house"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *HouseHandler) PublicDetail(c *gin.Context) {
	listingID, ok := bindDetailRequest(c)
	if !ok {
		return
	}

	userID, _ := minihandler.OptionalUserID(c)
	result, err := h.service.GetPublicDetail(c.Request.Context(), housesvc.DetailInput{ListingID: listingID, UserID: userID})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toDetailResponse(result))
}
