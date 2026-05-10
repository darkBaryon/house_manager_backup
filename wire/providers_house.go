package wire

import (
	househandler "house-manager/internal/handler/v1/house"
	housesvc "house-manager/internal/service/house"

	"github.com/google/wire"
)

var HouseSet = wire.NewSet(
	housesvc.NewHouseService,
	househandler.NewHouseHandler,
)
