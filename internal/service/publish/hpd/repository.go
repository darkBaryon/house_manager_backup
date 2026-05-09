package hpd

import (
	"context"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type hpdListingRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*model.HpdListing, error)
	FindBySource(ctx context.Context, sourceType model.HpdSourceType, sourceID bson.ObjectID) (*model.HpdListing, error)
	UpsertBySource(ctx context.Context, entity *model.HpdListing) (*model.HpdListing, error)
	UpdateLifecycleFields(ctx context.Context, id bson.ObjectID, fields bson.M) error
	UpdateStatus(ctx context.Context, id bson.ObjectID, listingStatus model.HpdListingStatus) error
}

type hpdMiniappListingRepository interface {
	UpsertByListingID(ctx context.Context, entity *model.HpdMiniappListing) (*model.HpdMiniappListing, error)
}

type hmdCentralizedRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error)
}

type hmdBuildingRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error)
}

type hmdDecentralizedRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdDecentralized, error)
}

type hmdRoomTypeCentralizedRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdRoomTypeCentralized, error)
}

type hmdRoomCentralizedRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error)
	ListByProjectID(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error)
	ListByBuildingID(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error)
	ListByRoomTypeID(ctx context.Context, roomTypeID bson.ObjectID) ([]model.HmdRoomCentralized, error)
}

type hmdRoomDecentralizedRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdRoomDecentralized, error)
	ListByDecentralizedID(ctx context.Context, decentralizedID bson.ObjectID) ([]model.HmdRoomDecentralized, error)
}
