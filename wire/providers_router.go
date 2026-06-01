package wire

import (
	"house-manager/internal/app"
	"house-manager/internal/config"
	"house-manager/internal/handler"
	"house-manager/internal/middleware"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func newInternalToolMiddleware(cfg *config.Config) ([]gin.HandlerFunc, error) {
	toolTimeout, err := cfg.AIChat.ToolTimeoutDuration()
	if err != nil {
		return nil, err
	}
	return []gin.HandlerFunc{
		middleware.InternalAuth(cfg.AIChat.InternalToken),
		middleware.RequestTimeout(toolTimeout),
	}, nil
}

func newRouteGroups(
	internalV1 internalV1Registrars,
	internalTools internalToolsRegistrars,
	internalToolMiddleware []gin.HandlerFunc,
	publicV1 publicV1Registrars,
	miniappProtectedV1 miniappProtectedV1Registrars,
	adminProtectedV1 adminProtectedV1Registrars,
	publishProtectedV1 publishProtectedV1Registrars,
	store *session.Store,
) []app.RouteGroup {
	miniappAuth := middleware.MiniappAuth(store)
	adminAuth := middleware.AdminAuth(store)
	publishAuth := middleware.PublishAuth(store)
	return []app.RouteGroup{
		{Prefix: "/api/v1", Registrars: []handler.RouteRegistrar(internalV1)},
		{Prefix: "/internal", Middleware: internalToolMiddleware, Registrars: []handler.RouteRegistrar(internalTools)},
		{Prefix: "/api/v1", Registrars: []handler.RouteRegistrar(publicV1)},
		{Prefix: "/api/v1", Middleware: []gin.HandlerFunc{miniappAuth}, Registrars: []handler.RouteRegistrar(miniappProtectedV1)},
		{Prefix: "/api/v1", Middleware: []gin.HandlerFunc{adminAuth}, Registrars: []handler.RouteRegistrar(adminProtectedV1)},
		{Prefix: "/api/v1", Middleware: []gin.HandlerFunc{publishAuth}, Registrars: []handler.RouteRegistrar(publishProtectedV1)},
	}
}

var RouterSet = wire.NewSet(
	newInternalV1Registrars,
	newInternalToolsRegistrars,
	newInternalToolMiddleware,
	newPublicV1Registrars,
	newMiniappProtectedV1Registrars,
	newAdminProtectedV1Registrars,
	newPublishProtectedV1Registrars,
	newRouteGroups,
)
