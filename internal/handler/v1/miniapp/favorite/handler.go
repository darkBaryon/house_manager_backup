package favorite

import (
	"context"

	"house-manager/internal/handler"
	favoritesvc "house-manager/internal/service/miniapp/favorite"
)

type Handler struct {
	service favoriteService
}

type favoriteService interface {
	Add(ctx context.Context, input favoritesvc.AddInput) (*favoritesvc.MutationResult, error)
	Remove(ctx context.Context, input favoritesvc.RemoveInput) (*favoritesvc.MutationResult, error)
	List(ctx context.Context, input favoritesvc.ListInput) (*favoritesvc.ListResult, error)
}

func NewHandler(service *favoritesvc.Service) *Handler {
	return newHandler(service)
}

func newHandler(service favoriteService) *Handler {
	return &Handler{service: service}
}

var _ handler.RouteRegistrar = (*Handler)(nil)
var _ favoriteService = (*favoritesvc.Service)(nil)
