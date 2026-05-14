package publish

import "github.com/gin-gonic/gin"

func (h *PublishHandler) RegisterRoutes(rg *gin.RouterGroup) {
	for _, registrar := range h.registrars {
		registrar.RegisterRoutes(rg)
	}
}
