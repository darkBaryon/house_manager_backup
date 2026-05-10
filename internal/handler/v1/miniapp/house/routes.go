package house

import "github.com/gin-gonic/gin"

func (h *HouseHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/house/search", h.Search)
	rg.POST("/house/public_detail", h.PublicDetail)
}
