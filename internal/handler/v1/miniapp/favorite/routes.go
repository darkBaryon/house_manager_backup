package favorite

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/favorite/add", h.Add)
	rg.POST("/favorite/remove", h.Remove)
	rg.POST("/favorite/list", h.List)
}
