package hpd

import (
	"context"
	"fmt"
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type EntrustRelationRepository struct {
	*common.Repository[hpdmodel.HpdEntrustRelation]
}

func NewEntrustRelationRepository(client *dbmongo.Client) *EntrustRelationRepository {
	return &EntrustRelationRepository{
		Repository: common.NewRepository[hpdmodel.HpdEntrustRelation](client.Collection(hpdmodel.CollectionHpdEntrustRelation)),
	}
}

func (r *EntrustRelationRepository) Create(ctx context.Context, entity *hpdmodel.HpdEntrustRelation) error {
	normalizeEntrustRelation(entity)
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create hpd entrust relation: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *EntrustRelationRepository) UpsertActiveByListingID(ctx context.Context, entity *hpdmodel.HpdEntrustRelation) (*hpdmodel.HpdEntrustRelation, error) {
	normalizeEntrustRelation(entity)
	if err := entity.ValidateForCreate(); err != nil {
		return nil, fmt.Errorf("upsert active hpd entrust relation by listingID: %w", err)
	}
	if entity.RelationStatus != hpdmodel.HpdRelationStatusActive {
		return nil, fmt.Errorf("upsert active hpd entrust relation by listingID: relationStatus must be active")
	}

	filter := activeFilter(bson.M{
		"listing_id":      entity.ListingID,
		"relation_status": hpdmodel.HpdRelationStatusActive,
	})
	if _, err := r.UpsertFields(ctx, filter, entrustRelationFields(entity)); err != nil {
		return nil, fmt.Errorf("upsert active hpd entrust relation by listingID: %w", err)
	}
	return r.FindActiveByListingID(ctx, entity.ListingID)
}

func (r *EntrustRelationRepository) FindActiveByListingID(ctx context.Context, listingID bson.ObjectID) (*hpdmodel.HpdEntrustRelation, error) {
	if listingID.IsZero() {
		return nil, fmt.Errorf("find active hpd entrust relation by listingID: listingID is required")
	}
	return r.FindOne(ctx, activeEntrustRelationFilter(bson.M{"listing_id": listingID}))
}

func (r *EntrustRelationRepository) ListActiveListingIDsByStaff(ctx context.Context, staffID bson.ObjectID) ([]bson.ObjectID, error) {
	if staffID.IsZero() {
		return nil, fmt.Errorf("list active listingIDs by staff: staffID is required")
	}
	filter := activeEntrustRelationFilter(bson.M{
		"$or": bson.A{
			bson.M{"maintainer_staff_id": staffID},
			bson.M{"service_staff_id": staffID},
		},
	})
	return r.listListingIDs(ctx, filter, "list active listingIDs by staff")
}

func (r *EntrustRelationRepository) ListActiveListingIDsByOwnerPhone(ctx context.Context, ownerPhone string) ([]bson.ObjectID, error) {
	ownerPhone = strings.TrimSpace(ownerPhone)
	if ownerPhone == "" {
		return nil, fmt.Errorf("list active listingIDs by owner phone: ownerPhone is required")
	}
	return r.listListingIDs(ctx, activeEntrustRelationFilter(bson.M{"owner_phone": ownerPhone}), "list active listingIDs by owner phone")
}

func (r *EntrustRelationRepository) CanAccessListing(ctx context.Context, listingID bson.ObjectID, staffID bson.ObjectID, ownerPhone string) (bool, error) {
	filter, err := activeEntrustAccessFilter(listingID, staffID, ownerPhone)
	if err != nil {
		return false, fmt.Errorf("can access hpd listing: %w", err)
	}
	total, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("can access hpd listing: %w", err)
	}
	return total > 0, nil
}

func (r *EntrustRelationRepository) listListingIDs(ctx context.Context, filter bson.M, action string) ([]bson.ObjectID, error) {
	rows, err := r.FindMany(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", action, err)
	}
	ids := make([]bson.ObjectID, 0, len(rows))
	seen := make(map[bson.ObjectID]struct{}, len(rows))
	for _, row := range rows {
		if row.ListingID.IsZero() {
			continue
		}
		if _, ok := seen[row.ListingID]; ok {
			continue
		}
		seen[row.ListingID] = struct{}{}
		ids = append(ids, row.ListingID)
	}
	return ids, nil
}

func normalizeEntrustRelation(entity *hpdmodel.HpdEntrustRelation) {
	if entity == nil {
		return
	}
	entity.OwnerName = strings.TrimSpace(entity.OwnerName)
	entity.OwnerPhone = strings.TrimSpace(entity.OwnerPhone)
	if entity.RelationStatus == hpdmodel.HpdRelationStatusUnspecified {
		entity.RelationStatus = hpdmodel.HpdRelationStatusActive
	}
}
