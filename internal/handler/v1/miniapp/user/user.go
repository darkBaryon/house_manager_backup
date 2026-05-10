package user

import (
	minihandler "house-manager/internal/handler/v1/miniapp/common"
	usersvc "house-manager/internal/service/miniapp/user"
	"house-manager/pkg/errcode"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Profile(c *gin.Context) {
	userID, ok := minihandler.CurrentUserID(c)
	if !ok {
		return
	}
	result, err := h.service.Profile(c.Request.Context(), usersvc.ProfileInput{UserID: userID})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toProfileResponse(result))
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, ok := minihandler.CurrentUserID(c)
	if !ok {
		return
	}
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(err))
		return
	}
	result, err := h.service.UpdateProfile(c.Request.Context(), usersvc.UpdateProfileInput{
		UserID:            userID,
		Nickname:          req.Nickname,
		Avatar:            req.Avatar,
		City:              req.City,
		BudgetMin:         req.BudgetMin,
		BudgetMax:         req.BudgetMax,
		PreferredAreas:    req.PreferredAreas,
		PreferredRentMode: req.PreferredRentMode,
		MoveInPlan:        req.MoveInPlan,
		Remark:            req.Remark,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toProfileResponse(result))
}

func (h *Handler) Dashboard(c *gin.Context) {
	userID, ok := minihandler.CurrentUserID(c)
	if !ok {
		return
	}
	result, err := h.service.Dashboard(c.Request.Context(), usersvc.DashboardInput{UserID: userID})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toDashboardResponse(result))
}
