package user

import (
	"context"

	"house-manager/internal/handler"
	usersvc "house-manager/internal/service/miniapp/user"
)

type Handler struct {
	service userService
}

type userService interface {
	Profile(ctx context.Context, input usersvc.ProfileInput) (*usersvc.Profile, error)
	UpdateProfile(ctx context.Context, input usersvc.UpdateProfileInput) (*usersvc.Profile, error)
	Dashboard(ctx context.Context, input usersvc.DashboardInput) (*usersvc.Dashboard, error)
}

func NewHandler(service *usersvc.Service) *Handler {
	return newHandler(service)
}

func newHandler(service userService) *Handler {
	return &Handler{service: service}
}

var _ handler.RouteRegistrar = (*Handler)(nil)
var _ userService = (*usersvc.Service)(nil)
