package wire

import (
	"house-manager/internal/app"
	"house-manager/internal/handler"
	"house-manager/internal/middleware"
	v1handler "house-manager/internal/handler/v1"
	v2handler "house-manager/internal/handler/v2"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
)

func newRouteGroups(
	roomH *v1handler.RoomHandler,
	healthH *v1handler.HealthHandler,
	v2RoomH *v2handler.RoomHandler,
	store *session.Store,
) []app.RouteGroup {
	auth := middleware.Auth(store)
	return []app.RouteGroup{
		{Prefix: "/api/v1", Registrars: []handler.RouteRegistrar{healthH}},
		{Prefix: "/api/v1", Middleware: []gin.HandlerFunc{auth}, Registrars: []handler.RouteRegistrar{roomH}},
		{Prefix: "/api/v2", Middleware: []gin.HandlerFunc{auth}, Registrars: []handler.RouteRegistrar{v2RoomH}},
	}
}
