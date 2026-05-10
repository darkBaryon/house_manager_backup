package wire

import (
	househandler "house-manager/internal/handler/v1/miniapp/house"
	housesvc "house-manager/internal/service/miniapp/house"

	"github.com/google/wire"
)

var MiniappHouseSet = wire.NewSet(
	housesvc.NewHouseService,
	househandler.NewHouseHandler,
)
