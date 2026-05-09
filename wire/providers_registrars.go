package wire

import (
	"house-manager/internal/handler"
	v1handler "house-manager/internal/handler/v1"
	publishhandler "house-manager/internal/handler/v1/publish"
)

type internalV1Registrars []handler.RouteRegistrar
type publicV1Registrars []handler.RouteRegistrar
type protectedV1Registrars []handler.RouteRegistrar

func newInternalV1Registrars(healthH *v1handler.HealthHandler) internalV1Registrars {
	return internalV1Registrars{healthH}
}

func newPublicV1Registrars(authH *v1handler.AuthHandler) publicV1Registrars {
	return publicV1Registrars{authH}
}

func newProtectedV1Registrars(sessionH *v1handler.SessionHandler, publishH *publishhandler.PublishHandler) protectedV1Registrars {
	return protectedV1Registrars{sessionH, publishH}
}
