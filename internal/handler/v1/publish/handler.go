package publish

import (
	"house-manager/internal/handler"
	publishsvc "house-manager/internal/service/publish"
)

type PublishHandler struct {
	service *publishsvc.PublishService
}

func NewPublishHandler(service *publishsvc.PublishService) *PublishHandler {
	return &PublishHandler{service: service}
}

var _ handler.RouteRegistrar = (*PublishHandler)(nil)
