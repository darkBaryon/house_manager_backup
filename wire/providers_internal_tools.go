package wire

import (
	internalhousehandler "house-manager/internal/handler/internaltools/house"
	housesvc "house-manager/internal/service/miniapp/house"

	"github.com/google/wire"
)

func newInternalToolsHouseHandler(service *housesvc.HouseService) *internalhousehandler.Handler {
	return internalhousehandler.NewHandler(service)
}

var InternalToolsSet = wire.NewSet(
	newInternalToolsHouseHandler,
)
