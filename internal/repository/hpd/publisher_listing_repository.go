package hpd

import (
	"context"
	"fmt"
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PublisherListingRepository struct {
	*common.Repository[hpdmodel.HpdPublisherListing]
}

func NewPublisherListingRepository(client *dbmongo.Client) *PublisherListingRepository {
	return &PublisherListingRepository{
		Repository: common.NewRepository[hpdmodel.HpdPublisherListing](client.Collection(hpdmodel.CollectionHpdPublisherListing)),
	}
}

func (r *PublisherListingRepository) Create(ctx context.Context, entity *hpdmodel.HpdPublisherListing) error {
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create hpd publisher listing: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *PublisherListingRepository) FindByID(ctx context.Context, id bson.ObjectID) (*hpdmodel.HpdPublisherListing, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hpd publisher listing by id: id is required")
	}
	return r.Repository.FindByID(ctx, id)
}

func (r *PublisherListingRepository) FindByListingID(ctx context.Context, listingID bson.ObjectID) (*hpdmodel.HpdPublisherListing, error) {
	if listingID.IsZero() {
		return nil, fmt.Errorf("find hpd publisher listing by listingID: listingID is required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"listing_id": listingID}))
}

func (r *PublisherListingRepository) FindBySource(ctx context.Context, sourceType hpdmodel.HpdSourceType, sourceID bson.ObjectID) (*hpdmodel.HpdPublisherListing, error) {
	if !sourceType.Valid() || sourceID.IsZero() {
		return nil, fmt.Errorf("find hpd publisher listing by source: sourceType and sourceID are required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"source_type": sourceType, "source_id": sourceID}))
}

func (r *PublisherListingRepository) UpsertByListingID(ctx context.Context, entity *hpdmodel.HpdPublisherListing) (*hpdmodel.HpdPublisherListing, error) {
	if err := entity.ValidateForCreate(); err != nil {
		return nil, fmt.Errorf("upsert hpd publisher listing by listingID: %w", err)
	}
	filter := activeFilter(bson.M{"listing_id": entity.ListingID})
	if _, err := r.UpsertFields(ctx, filter, publisherListingFields(entity)); err != nil {
		return nil, fmt.Errorf("upsert hpd publisher listing by listingID: %w", err)
	}
	return r.FindByListingID(ctx, entity.ListingID)
}

func (r *PublisherListingRepository) UpdateProjectionFields(ctx context.Context, listingID bson.ObjectID, fields bson.M) error {
	if listingID.IsZero() {
		return fmt.Errorf("update hpd publisher listing projection fields: listingID is required")
	}
	safeFields, err := pickAllowedFields(fields, publisherProjectionFields)
	if err != nil {
		return fmt.Errorf("update hpd publisher listing projection fields: %w", err)
	}
	if err := hpdmodel.ValidateHpdUpdateFields(safeFields); err != nil {
		return fmt.Errorf("update hpd publisher listing projection fields: %w", err)
	}

	setFields := cloneBsonM(safeFields)
	setFields["updated_at"] = time.Now().Unix()
	update := bson.M{
		"$set": setFields,
		"$inc": bson.M{"version": 1},
	}
	res, err := r.Collection.UpdateOne(ctx, activeFilter(bson.M{"listing_id": listingID}), update)
	if err != nil {
		return fmt.Errorf("update hpd publisher listing projection fields: %w", err)
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *PublisherListingRepository) ListByRoot(ctx context.Context, rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID) ([]hpdmodel.HpdPublisherListing, error) {
	if !rootType.Valid() {
		return nil, fmt.Errorf("list hpd publisher listings by root: rootType is invalid")
	}
	if rootID.IsZero() {
		return nil, fmt.Errorf("list hpd publisher listings by root: rootID is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{
		"root_type": rootType,
		"root_id":   rootID,
	}))
}
