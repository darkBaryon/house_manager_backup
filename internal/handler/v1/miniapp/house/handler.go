package house

import (
	"context"

	"house-manager/internal/handler"
	housesvc "house-manager/internal/service/miniapp/house"
)

type HouseHandler struct {
	service houseService
}

type houseService interface {
	Search(ctx context.Context, input housesvc.SearchInput) (*housesvc.SearchResult, error)
	GetPublicDetail(ctx context.Context, input housesvc.DetailInput) (*housesvc.DetailResult, error)
}

func NewHouseHandler(service *housesvc.HouseService) *HouseHandler {
	return newHouseHandler(service)
}

func newHouseHandler(service houseService) *HouseHandler {
	return &HouseHandler{service: service}
}

var _ handler.RouteRegistrar = (*HouseHandler)(nil)
var _ houseService = (*housesvc.HouseService)(nil)
