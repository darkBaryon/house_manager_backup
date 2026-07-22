package house

import (
	"context"

	"house-manager/internal/handler"
	housesvc "house-manager/internal/service/miniapp/house"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service houseService
}

type houseService interface {
	Search(ctx context.Context, input housesvc.SearchInput) (*housesvc.SearchResult, error)
	GetPublicDetail(ctx context.Context, input housesvc.DetailInput) (*housesvc.DetailResult, error)
}

func NewHandler(service *housesvc.HouseService) *Handler {
	return newHandler(service)
}

func newHandler(service houseService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/tools/house/search", h.Search)
	rg.POST("/tools/house/public_detail", h.PublicDetail)
}

var _ handler.RouteRegistrar = (*Handler)(nil)
var _ houseService = (*housesvc.HouseService)(nil)
