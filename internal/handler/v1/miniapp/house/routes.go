package house

import (
	"house-manager/internal/middleware"

	"github.com/gin-gonic/gin"
)

func (h *HouseHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/house/search", h.Search)
	rg.POST("/house/public_detail", middleware.OptionalAuth(h.sessionStore), h.PublicDetail)
}
