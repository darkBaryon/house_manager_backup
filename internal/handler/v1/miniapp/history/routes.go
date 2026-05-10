package history

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/history/add", h.Add)
	rg.POST("/history/list", h.List)
}
