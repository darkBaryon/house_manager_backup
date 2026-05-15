package publishaccess

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hpdmodel "house-manager/internal/model/hpd"
)

type hpdListingRepository interface {
	FindBySource(ctx context.Context, sourceType hpdmodel.HpdSourceType, sourceID bson.ObjectID) (*hpdmodel.HpdListing, error)
	ListByIDs(ctx context.Context, ids []bson.ObjectID) ([]hpdmodel.HpdListing, error)
}

type hpdEntrustRelationRepository interface {
	UpsertActiveByListingID(ctx context.Context, entity *hpdmodel.HpdEntrustRelation) (*hpdmodel.HpdEntrustRelation, error)
	FindActiveByListingID(ctx context.Context, listingID bson.ObjectID) (*hpdmodel.HpdEntrustRelation, error)
	ListActiveListingIDsByStaff(ctx context.Context, staffID bson.ObjectID) ([]bson.ObjectID, error)
	ListActiveListingIDsByOwnerPhone(ctx context.Context, ownerPhone string) ([]bson.ObjectID, error)
	CanAccessListing(ctx context.Context, listingID, staffID bson.ObjectID, ownerPhone string) (bool, error)
}
