package favorite

import (
	minihandler "house-manager/internal/handler/v1/miniapp/common"
	favoritesvc "house-manager/internal/service/miniapp/favorite"
	"house-manager/pkg/errcode"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Add(c *gin.Context) {
	userID, ok := minihandler.CurrentUserID(c)
	if !ok {
		return
	}
	var req listingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(err))
		return
	}
	listingID, ok := minihandler.ParseObjectID(c, req.ListingID)
	if !ok {
		return
	}
	result, err := h.service.Add(c.Request.Context(), favoritesvc.AddInput{UserID: userID, ListingID: listingID})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toMutationResponse(result))
}

func (h *Handler) Remove(c *gin.Context) {
	userID, ok := minihandler.CurrentUserID(c)
	if !ok {
		return
	}
	var req listingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(err))
		return
	}
	listingID, ok := minihandler.ParseObjectID(c, req.ListingID)
	if !ok {
		return
	}
	result, err := h.service.Remove(c.Request.Context(), favoritesvc.RemoveInput{UserID: userID, ListingID: listingID})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toMutationResponse(result))
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := minihandler.CurrentUserID(c)
	if !ok {
		return
	}
	var req listRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(err))
		return
	}
	result, err := h.service.List(c.Request.Context(), favoritesvc.ListInput{UserID: userID, Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toListResponse(result))
}
