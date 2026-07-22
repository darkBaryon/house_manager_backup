package wire

import (
	"house-manager/internal/handler"
	internalhousehandler "house-manager/internal/handler/internaltools/house"
	v1handler "house-manager/internal/handler/v1"
	adminauthhandler "house-manager/internal/handler/v1/admin/auth"
	adminhousehandler "house-manager/internal/handler/v1/admin/house"
	adminproviderhandler "house-manager/internal/handler/v1/admin/provider"
	adminrolehandler "house-manager/internal/handler/v1/admin/role"
	adminstaffhandler "house-manager/internal/handler/v1/admin/staff"
	authhandler "house-manager/internal/handler/v1/miniapp/auth"
	chathandler "house-manager/internal/handler/v1/miniapp/chat"
	favoritehandler "house-manager/internal/handler/v1/miniapp/favorite"
	historyhandler "house-manager/internal/handler/v1/miniapp/history"
	househandler "house-manager/internal/handler/v1/miniapp/house"
	userhandler "house-manager/internal/handler/v1/miniapp/user"
	publishhandler "house-manager/internal/handler/v1/publish"
	publishauthhandler "house-manager/internal/handler/v1/publish/auth"
)

type internalV1Registrars []handler.RouteRegistrar
type internalToolsRegistrars []handler.RouteRegistrar
type publicV1Registrars []handler.RouteRegistrar
type miniappProtectedV1Registrars []handler.RouteRegistrar
type adminProtectedV1Registrars []handler.RouteRegistrar
type publishProtectedV1Registrars []handler.RouteRegistrar

func newInternalV1Registrars(healthH *v1handler.HealthHandler) internalV1Registrars {
	return internalV1Registrars{healthH}
}

func newInternalToolsRegistrars(houseToolsH *internalhousehandler.Handler) internalToolsRegistrars {
	return internalToolsRegistrars{houseToolsH}
}

func newPublicV1Registrars(
	authH *authhandler.AuthHandler,
	houseH *househandler.HouseHandler,
	adminAuthH *adminauthhandler.PublicHandler,
	publishAuthH *publishauthhandler.PublicHandler,
) publicV1Registrars {
	return publicV1Registrars{authH, houseH, adminAuthH, publishAuthH}
}

func newMiniappProtectedV1Registrars(
	sessionH *authhandler.SessionHandler,
	chatH *chathandler.Handler,
	favoriteH *favoritehandler.Handler,
	historyH *historyhandler.Handler,
	userH *userhandler.Handler,
) miniappProtectedV1Registrars {
	return miniappProtectedV1Registrars{sessionH, chatH, favoriteH, historyH, userH}
}

func newAdminProtectedV1Registrars(
	adminAuthH *adminauthhandler.Handler,
	adminStaffH *adminstaffhandler.Handler,
	adminRoleH *adminrolehandler.Handler,
	adminProviderH *adminproviderhandler.Handler,
	adminHouseH *adminhousehandler.Handler,
) adminProtectedV1Registrars {
	return adminProtectedV1Registrars{adminAuthH, adminStaffH, adminRoleH, adminProviderH, adminHouseH}
}

func newPublishProtectedV1Registrars(
	publishAuthH *publishauthhandler.Handler,
	publishH *publishhandler.PublishHandler,
) publishProtectedV1Registrars {
	return publishProtectedV1Registrars{publishAuthH, publishH}
}
