package hpd

import (
	"context"
	"fmt"
	"time"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ListingRepository struct {
	*common.Repository[model.HpdListing]
}

func NewListingRepository(client *dbmongo.Client) *ListingRepository {
	return &ListingRepository{
		Repository: common.NewRepository[model.HpdListing](client.Collection(model.CollectionHpdListing)),
	}
}

func (r *ListingRepository) Create(ctx context.Context, entity *model.HpdListing) error {
	if entity != nil && entity.ListingStatus == model.HpdListingStatusUnspecified {
		entity.ListingStatus = model.HpdListingStatusDraft
	}
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create hpd listing: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *ListingRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.HpdListing, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hpd listing by id: id is required")
	}
	return r.Repository.FindByID(ctx, id)
}

func (r *ListingRepository) FindBySource(ctx context.Context, sourceType model.HpdSourceType, sourceID bson.ObjectID) (*model.HpdListing, error) {
	if !sourceType.Valid() || sourceID.IsZero() {
		return nil, fmt.Errorf("find hpd listing by source: sourceType and sourceID are required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"source_type": sourceType, "source_id": sourceID}))
}

func (r *ListingRepository) UpsertBySource(ctx context.Context, entity *model.HpdListing) (*model.HpdListing, error) {
	if entity != nil && entity.ListingStatus == model.HpdListingStatusUnspecified {
		entity.ListingStatus = model.HpdListingStatusDraft
	}
	if err := entity.ValidateForCreate(); err != nil {
		return nil, fmt.Errorf("upsert hpd listing by source: %w", err)
	}
	filter := activeFilter(bson.M{"source_type": entity.SourceType, "source_id": entity.SourceID})
	fields := listingFields(entity)

	now := time.Now().Unix()
	fields["updated_at"] = now
	update := bson.M{
		"$set": fields,
		"$inc": bson.M{"version": 1},
		"$setOnInsert": bson.M{
			"created_at":     now,
			"status":         model.StatusActive,
			"listing_status": entity.ListingStatus,
			"published_at":   entity.PublishedAt,
			"offline_at":     entity.OfflineAt,
		},
	}
	opts := options.UpdateOne().SetUpsert(true)
	if _, err := r.Collection.UpdateOne(ctx, filter, update, opts); err != nil {
		return nil, fmt.Errorf("upsert hpd listing by source: %w", err)
	}
	return r.FindBySource(ctx, entity.SourceType, entity.SourceID)
}

func (r *ListingRepository) UpdateLifecycleFields(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hpd listing lifecycle fields: id is required")
	}
	safeFields, err := pickAllowedFields(fields, listingLifecycleFields)
	if err != nil {
		return fmt.Errorf("update hpd listing lifecycle fields: %w", err)
	}
	if err := model.ValidateHpdUpdateFields(safeFields); err != nil {
		return fmt.Errorf("update hpd listing lifecycle fields: %w", err)
	}
	return r.UpdateFieldsByID(ctx, id, safeFields)
}

func (r *ListingRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, listingStatus model.HpdListingStatus) error {
	if id.IsZero() {
		return fmt.Errorf("update hpd listing status: id is required")
	}
	if listingStatus == model.HpdListingStatusUnspecified || !listingStatus.Valid() {
		return fmt.Errorf("update hpd listing status: listingStatus is invalid")
	}
	return r.UpdateFieldsByID(ctx, id, listingStatusUpdateFields(listingStatus))
}
