package history

import (
	minihandler "house-manager/internal/handler/v1/miniapp/common"
	historysvc "house-manager/internal/service/miniapp/history"
	"house-manager/pkg/errcode"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Add(c *gin.Context) {
	userID, ok := minihandler.CurrentUserID(c)
	if !ok {
		return
	}
	var req addRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(err))
		return
	}
	listingID, ok := minihandler.ParseObjectID(c, req.ListingID)
	if !ok {
		return
	}
	result, err := h.service.Add(c.Request.Context(), historysvc.AddInput{UserID: userID, ListingID: listingID, Source: req.Source})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toAddResponse(result))
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
	result, err := h.service.List(c.Request.Context(), historysvc.ListInput{UserID: userID, Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toListResponse(result))
}
