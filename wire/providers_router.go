package wire

import (
	"house-manager/internal/app"
	"house-manager/internal/handler"
	"house-manager/internal/middleware"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func newRouteGroups(
	internalV1 internalV1Registrars,
	publicV1 publicV1Registrars,
	protectedV1 protectedV1Registrars,
	store *session.Store,
) []app.RouteGroup {
	auth := middleware.Auth(store)
	return []app.RouteGroup{
		{Prefix: "/api/v1", Registrars: []handler.RouteRegistrar(internalV1)},
		{Prefix: "/api/v1", Registrars: []handler.RouteRegistrar(publicV1)},
		{Prefix: "/api/v1", Middleware: []gin.HandlerFunc{auth}, Registrars: []handler.RouteRegistrar(protectedV1)},
	}
}

var RouterSet = wire.NewSet(
	newInternalV1Registrars,
	newPublicV1Registrars,
	newProtectedV1Registrars,
	newRouteGroups,
)
