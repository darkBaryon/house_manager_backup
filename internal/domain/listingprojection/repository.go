package listingprojection

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
)

type hpdListingRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*hpdmodel.HpdListing, error)
	UpsertBySource(ctx context.Context, entity *hpdmodel.HpdListing) (*hpdmodel.HpdListing, error)
	UpdateLifecycleFields(ctx context.Context, id bson.ObjectID, fields bson.M) error
	UpdateStatus(ctx context.Context, id bson.ObjectID, listingStatus hpdmodel.HpdListingStatus) error
}

type hpdMiniappListingRepository interface {
	UpsertByListingID(ctx context.Context, entity *hpdmodel.HpdMiniappListing) (*hpdmodel.HpdMiniappListing, error)
}

type hpdPublisherListingRepository interface {
	UpsertByListingID(ctx context.Context, entity *hpdmodel.HpdPublisherListing) (*hpdmodel.HpdPublisherListing, error)
}

type hpdRootScopeRepository interface {
	FindActiveByRoot(ctx context.Context, rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID) (*hpdmodel.HpdRootScopeRelation, error)
}

type hmdCentralizedRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdCentralized, error)
}

type hmdBuildingRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdBuilding, error)
}

type hmdDecentralizedRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdDecentralized, error)
}

type hmdRoomTypeCentralizedRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomTypeCentralized, error)
}

type hmdRoomCentralizedRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomCentralized, error)
	ListByProjectID(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error)
	ListByBuildingID(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error)
	ListByRoomTypeID(ctx context.Context, roomTypeID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error)
}

type hmdRoomDecentralizedRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomDecentralized, error)
	ListByDecentralizedID(ctx context.Context, decentralizedID bson.ObjectID) ([]hmdmodel.HmdRoomDecentralized, error)
}
