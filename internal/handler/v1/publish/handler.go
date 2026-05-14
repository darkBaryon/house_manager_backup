package publish

import (
	"house-manager/internal/handler"
	"house-manager/internal/handler/v1/publish/building"
	"house-manager/internal/handler/v1/publish/community"
	"house-manager/internal/handler/v1/publish/project"
	"house-manager/internal/handler/v1/publish/room"
	"house-manager/internal/handler/v1/publish/roomtype"
	publishsvc "house-manager/internal/service/publish"
)

type publishService interface {
	project.Service
	building.Service
	roomtype.Service
	room.Service
	community.Service
}

type PublishHandler struct {
	registrars []handler.RouteRegistrar
}

func NewPublishHandler(service *publishsvc.PublishService) *PublishHandler {
	return newPublishHandler(service)
}

func newPublishHandler(service publishService) *PublishHandler {
	return &PublishHandler{
		registrars: []handler.RouteRegistrar{
			project.NewHandler(service),
			building.NewHandler(service),
			roomtype.NewHandler(service),
			room.NewHandler(service),
			community.NewHandler(service),
		},
	}
}

var _ handler.RouteRegistrar = (*PublishHandler)(nil)
var _ publishService = (*publishsvc.PublishService)(nil)
