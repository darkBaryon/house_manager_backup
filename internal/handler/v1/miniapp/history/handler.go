package history

import (
	"context"

	"house-manager/internal/handler"
	historysvc "house-manager/internal/service/miniapp/history"
)

type Handler struct {
	service historyService
}

type historyService interface {
	Add(ctx context.Context, input historysvc.AddInput) (*historysvc.AddResult, error)
	List(ctx context.Context, input historysvc.ListInput) (*historysvc.ListResult, error)
}

func NewHandler(service *historysvc.Service) *Handler {
	return newHandler(service)
}

func newHandler(service historyService) *Handler {
	return &Handler{service: service}
}

var _ handler.RouteRegistrar = (*Handler)(nil)
var _ historyService = (*historysvc.Service)(nil)
