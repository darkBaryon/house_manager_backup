package user

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/user/profile", h.Profile)
	rg.POST("/user/update_profile", h.UpdateProfile)
	rg.POST("/user/dashboard", h.Dashboard)
}
