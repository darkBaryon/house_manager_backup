package wire

import (
	"house-manager/internal/handler"
	v1handler "house-manager/internal/handler/v1"
)

type internalV1Registrars []handler.RouteRegistrar
type publicV1Registrars []handler.RouteRegistrar
type protectedV1Registrars []handler.RouteRegistrar

func newInternalV1Registrars(healthH *v1handler.HealthHandler) internalV1Registrars {
	return internalV1Registrars{healthH}
}

func newPublicV1Registrars() publicV1Registrars {
	return publicV1Registrars{}
}

func newProtectedV1Registrars(roomH *v1handler.RoomHandler) protectedV1Registrars {
	return protectedV1Registrars{roomH}
}
