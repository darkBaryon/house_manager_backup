package house

import (
	"context"

	"house-manager/internal/handler"
	housesvc "house-manager/internal/service/miniapp/house"
	"house-manager/pkg/session"
)

type HouseHandler struct {
	service      houseService
	sessionStore *session.Store
}

type houseService interface {
	Search(ctx context.Context, input housesvc.SearchInput) (*housesvc.SearchResult, error)
	GetPublicDetail(ctx context.Context, input housesvc.DetailInput) (*housesvc.DetailResult, error)
}

func NewHouseHandler(service *housesvc.HouseService, store *session.Store) *HouseHandler {
	h := newHouseHandler(service)
	h.sessionStore = store
	return h
}

func newHouseHandler(service houseService) *HouseHandler {
	return &HouseHandler{service: service}
}

var _ handler.RouteRegistrar = (*HouseHandler)(nil)
var _ houseService = (*housesvc.HouseService)(nil)
