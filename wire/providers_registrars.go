package wire

import (
	"house-manager/internal/handler"
	v1handler "house-manager/internal/handler/v1"
	authhandler "house-manager/internal/handler/v1/miniapp/auth"
	favoritehandler "house-manager/internal/handler/v1/miniapp/favorite"
	historyhandler "house-manager/internal/handler/v1/miniapp/history"
	househandler "house-manager/internal/handler/v1/miniapp/house"
	userhandler "house-manager/internal/handler/v1/miniapp/user"
	publishhandler "house-manager/internal/handler/v1/publish"
	publishauthhandler "house-manager/internal/handler/v1/publish/auth"
)

type internalV1Registrars []handler.RouteRegistrar
type publicV1Registrars []handler.RouteRegistrar
type miniappProtectedV1Registrars []handler.RouteRegistrar
type publishProtectedV1Registrars []handler.RouteRegistrar

func newInternalV1Registrars(healthH *v1handler.HealthHandler) internalV1Registrars {
	return internalV1Registrars{healthH}
}

func newPublicV1Registrars(
	authH *authhandler.AuthHandler,
	houseH *househandler.HouseHandler,
	publishAuthH *publishauthhandler.PublicHandler,
) publicV1Registrars {
	return publicV1Registrars{authH, houseH, publishAuthH}
}

func newMiniappProtectedV1Registrars(
	sessionH *authhandler.SessionHandler,
	favoriteH *favoritehandler.Handler,
	historyH *historyhandler.Handler,
	userH *userhandler.Handler,
) miniappProtectedV1Registrars {
	return miniappProtectedV1Registrars{sessionH, favoriteH, historyH, userH}
}

func newPublishProtectedV1Registrars(
	publishAuthH *publishauthhandler.Handler,
	publishH *publishhandler.PublishHandler,
) publishProtectedV1Registrars {
	return publishProtectedV1Registrars{publishAuthH, publishH}
}
