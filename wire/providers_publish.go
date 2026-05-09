package wire

import (
	publishhandler "house-manager/internal/handler/v1/publish"
	repohmd "house-manager/internal/repository/hmd"
	repohpd "house-manager/internal/repository/hpd"
	hmdsvc "house-manager/internal/service/hmd"
	hpdsvc "house-manager/internal/service/hpd"
	publishsvc "house-manager/internal/service/publish"
	dbmongo "house-manager/pkg/database/mongo"

	"github.com/google/wire"
)

func newHmdCentralizedRepository(client *dbmongo.Client) *repohmd.CentralizedRepository {
	return repohmd.NewCentralizedRepository(client)
}

func newHmdBuildingRepository(client *dbmongo.Client) *repohmd.BuildingRepository {
	return repohmd.NewBuildingRepository(client)
}

func newHmdDecentralizedRepository(client *dbmongo.Client) *repohmd.DecentralizedRepository {
	return repohmd.NewDecentralizedRepository(client)
}

func newHmdRoomTypeCentralizedRepository(client *dbmongo.Client) *repohmd.RoomTypeCentralizedRepository {
	return repohmd.NewRoomTypeCentralizedRepository(client)
}

func newHmdRoomCentralizedRepository(client *dbmongo.Client) *repohmd.RoomCentralizedRepository {
	return repohmd.NewRoomCentralizedRepository(client)
}

func newHmdRoomDecentralizedRepository(client *dbmongo.Client) *repohmd.RoomDecentralizedRepository {
	return repohmd.NewRoomDecentralizedRepository(client)
}

func newHpdListingRepository(client *dbmongo.Client) *repohpd.ListingRepository {
	return repohpd.NewListingRepository(client)
}

func newHpdMiniappListingRepository(client *dbmongo.Client) *repohpd.MiniappListingRepository {
	return repohpd.NewMiniappListingRepository(client)
}

var HmdSet = wire.NewSet(
	newHmdCentralizedRepository,
	newHmdBuildingRepository,
	newHmdDecentralizedRepository,
	newHmdRoomTypeCentralizedRepository,
	newHmdRoomCentralizedRepository,
	newHmdRoomDecentralizedRepository,
	hmdsvc.NewService,
)

var HpdSet = wire.NewSet(
	newHpdListingRepository,
	newHpdMiniappListingRepository,
	hpdsvc.NewService,
)

var PublishSet = wire.NewSet(
	HmdSet,
	HpdSet,
	publishsvc.NewPublishService,
	publishhandler.NewPublishHandler,
)
