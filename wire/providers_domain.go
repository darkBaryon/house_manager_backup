package wire

import (
	"context"

	hmddomain "house-manager/internal/domain/hmd"
	"house-manager/internal/domain/listingprojection"
	"house-manager/internal/domain/publishaccess"
	repohmd "house-manager/internal/repository/hmd"
	repohpd "house-manager/internal/repository/hpd"
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

func newHpdPublisherListingRepository(ctx context.Context, client *dbmongo.Client) (*repohpd.PublisherListingRepository, error) {
	repo := repohpd.NewPublisherListingRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newHpdRootScopeRepository(ctx context.Context, client *dbmongo.Client) (*repohpd.RootScopeRepository, error) {
	repo := repohpd.NewRootScopeRepository(client)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func newHpdMiniappProjector(
	hpdListingRepo *repohpd.ListingRepository,
	hpdMiniappListingRepo *repohpd.MiniappListingRepository,
	hmdCentralizedRepo *repohmd.CentralizedRepository,
	hmdBuildingRepo *repohmd.BuildingRepository,
	hmdDecentralizedRepo *repohmd.DecentralizedRepository,
	hmdRoomTypeCentralizedRepo *repohmd.RoomTypeCentralizedRepository,
	hmdRoomCentralizedRepo *repohmd.RoomCentralizedRepository,
	hmdRoomDecentralizedRepo *repohmd.RoomDecentralizedRepository,
) *listingprojection.MiniappProjector {
	return listingprojection.NewMiniappProjector(
		hpdListingRepo,
		hpdMiniappListingRepo,
		hmdCentralizedRepo,
		hmdBuildingRepo,
		hmdDecentralizedRepo,
		hmdRoomTypeCentralizedRepo,
		hmdRoomCentralizedRepo,
		hmdRoomDecentralizedRepo,
	)
}

func newHpdPublisherProjector(
	hpdListingRepo *repohpd.ListingRepository,
	hpdPublisherListingRepo *repohpd.PublisherListingRepository,
	hmdCentralizedRepo *repohmd.CentralizedRepository,
	hmdBuildingRepo *repohmd.BuildingRepository,
	hmdDecentralizedRepo *repohmd.DecentralizedRepository,
	hmdRoomTypeCentralizedRepo *repohmd.RoomTypeCentralizedRepository,
	hmdRoomCentralizedRepo *repohmd.RoomCentralizedRepository,
	hmdRoomDecentralizedRepo *repohmd.RoomDecentralizedRepository,
) *listingprojection.PublisherProjector {
	return listingprojection.NewPublisherProjector(
		hpdListingRepo,
		hpdPublisherListingRepo,
		hmdCentralizedRepo,
		hmdBuildingRepo,
		hmdDecentralizedRepo,
		hmdRoomTypeCentralizedRepo,
		hmdRoomCentralizedRepo,
		hmdRoomDecentralizedRepo,
	)
}

var DomainHmdSet = wire.NewSet(
	newHmdCentralizedRepository,
	newHmdBuildingRepository,
	newHmdDecentralizedRepository,
	newHmdRoomTypeCentralizedRepository,
	newHmdRoomCentralizedRepository,
	newHmdRoomDecentralizedRepository,
	hmddomain.NewService,
)

var HpdStorageSet = wire.NewSet(
	newHpdListingRepository,
	newHpdMiniappListingRepository,
	newHpdPublisherListingRepository,
	newHpdRootScopeRepository,
)

var ListingProjectionSet = wire.NewSet(
	newHpdMiniappProjector,
	newHpdPublisherProjector,
	listingprojection.NewService,
)

var PublishAccessSet = wire.NewSet(
	publishaccess.NewService,
)
