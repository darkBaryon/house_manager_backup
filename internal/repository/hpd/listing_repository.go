package hpd

import (
	"context"
	"fmt"

	commonmodel "house-manager/internal/model/common"
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ListingRepository struct {
	*common.Repository[hpdmodel.HpdListing]
}

func NewListingRepository(client *dbmongo.Client) *ListingRepository {
	return &ListingRepository{
		Repository: common.NewRepository[hpdmodel.HpdListing](client.Collection(hpdmodel.CollectionHpdListing)),
	}
}

func (r *ListingRepository) Create(ctx context.Context, entity *hpdmodel.HpdListing) error {
	if entity != nil && entity.ListingStatus == hpdmodel.HpdListingStatusUnspecified {
		entity.ListingStatus = hpdmodel.HpdListingStatusDraft
	}
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create hpd listing: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *ListingRepository) FindByID(ctx context.Context, id bson.ObjectID) (*hpdmodel.HpdListing, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hpd listing by id: id is required")
	}
	return r.Repository.FindByID(ctx, id)
}

func (r *ListingRepository) FindBySource(ctx context.Context, sourceType hpdmodel.HpdSourceType, sourceID bson.ObjectID) (*hpdmodel.HpdListing, error) {
	if !sourceType.Valid() || sourceID.IsZero() {
		return nil, fmt.Errorf("find hpd listing by source: sourceType and sourceID are required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"source_type": sourceType, "source_id": sourceID}))
}

func (r *ListingRepository) ListByIDs(ctx context.Context, ids []bson.ObjectID) ([]hpdmodel.HpdListing, error) {
	ids = compactObjectIDs(ids)
	if len(ids) == 0 {
		return []hpdmodel.HpdListing{}, nil
	}
	return r.FindMany(ctx, activeFilter(bson.M{"_id": bson.M{"$in": ids}}))
}

func (r *ListingRepository) UpsertBySource(ctx context.Context, entity *hpdmodel.HpdListing) (*hpdmodel.HpdListing, error) {
	if entity != nil && entity.ListingStatus == hpdmodel.HpdListingStatusUnspecified {
		entity.ListingStatus = hpdmodel.HpdListingStatusDraft
	}
	if err := entity.ValidateForCreate(); err != nil {
		return nil, fmt.Errorf("upsert hpd listing by source: %w", err)
	}
	filter := common.And(
		common.Eq(hpdFieldSourceType, entity.SourceType),
		common.Eq(hpdFieldSourceID, entity.SourceID),
		common.Active(),
	)
	fields := listingFields(entity)

	now := time.Now().Unix()
	fields["updated_at"] = now
	update := updateDocFromSetFields(fields).
		SetOnInsert(hpdFieldCreatedAt, now).
		SetOnInsert(hpdFieldStatus, commonmodel.StatusActive).
		SetOnInsert(hpdFieldListingStatus, entity.ListingStatus).
		SetOnInsert(common.Field("published_at"), entity.PublishedAt).
		SetOnInsert(common.Field("offline_at"), entity.OfflineAt)
	if _, err := r.UpsertOneBy(ctx, filter, update); err != nil {
		return nil, fmt.Errorf("upsert hpd listing by source: %w", err)
	}
	return r.FindBySource(ctx, entity.SourceType, entity.SourceID)
}

func compactObjectIDs(ids []bson.ObjectID) []bson.ObjectID {
	compacted := make([]bson.ObjectID, 0, len(ids))
	seen := make(map[bson.ObjectID]struct{}, len(ids))
	for _, id := range ids {
		if id.IsZero() {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		compacted = append(compacted, id)
	}
	return compacted
}
