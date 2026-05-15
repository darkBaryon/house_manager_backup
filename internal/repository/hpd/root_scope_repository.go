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

type RootScopeRepository struct {
	*common.Repository[hpdmodel.HpdRootScopeRelation]
}

func NewRootScopeRepository(client *dbmongo.Client) *RootScopeRepository {
	return &RootScopeRepository{
		Repository: common.NewRepository[hpdmodel.HpdRootScopeRelation](client.Collection(hpdmodel.CollectionHpdRootScope)),
	}
}

func (r *RootScopeRepository) Create(ctx context.Context, entity *hpdmodel.HpdRootScopeRelation) error {
	normalizeRootScopeRelation(entity)
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create hpd root scope relation: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *RootScopeRepository) UpsertActiveByRoot(ctx context.Context, entity *hpdmodel.HpdRootScopeRelation) (*hpdmodel.HpdRootScopeRelation, error) {
	normalizeRootScopeRelation(entity)
	if err := entity.ValidateForCreate(); err != nil {
		return nil, fmt.Errorf("upsert active hpd root scope relation by root: %w", err)
	}
	if entity.RelationStatus != hpdmodel.HpdRelationStatusActive {
		return nil, fmt.Errorf("upsert active hpd root scope relation by root: relationStatus must be active")
	}
	filter := activeRootScopeRelationFilter(bson.M{
		"root_type":       entity.RootType,
		"root_id":         entity.RootID,
		"owner_phone":     entity.OwnerPhone,
		"relation_status": hpdmodel.HpdRelationStatusActive,
	})
	if _, err := r.UpsertFields(ctx, filter, rootScopeRelationFields(entity)); err != nil {
		return nil, fmt.Errorf("upsert active hpd root scope relation by root: %w", err)
	}
	return r.FindActiveByRoot(ctx, entity.RootType, entity.RootID, entity.OwnerPhone)
}

func (r *RootScopeRepository) FindActiveByRoot(ctx context.Context, rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, ownerPhone string) (*hpdmodel.HpdRootScopeRelation, error) {
	ownerPhone = strings.TrimSpace(ownerPhone)
	if !rootType.Valid() {
		return nil, fmt.Errorf("find active hpd root scope relation by root: rootType is invalid")
	}
	if rootID.IsZero() {
		return nil, fmt.Errorf("find active hpd root scope relation by root: rootID is required")
	}
	if ownerPhone == "" {
		return nil, fmt.Errorf("find active hpd root scope relation by root: ownerPhone is required")
	}
	return r.FindOne(ctx, activeRootScopeRelationFilter(bson.M{
		"root_type":   rootType,
		"root_id":     rootID,
		"owner_phone": ownerPhone,
	}))
}

func (r *RootScopeRepository) ListActiveRootIDsByOwnerPhone(ctx context.Context, rootType hpdmodel.HpdRootScopeType, ownerPhone string) ([]bson.ObjectID, error) {
	ownerPhone = strings.TrimSpace(ownerPhone)
	if !rootType.Valid() {
		return nil, fmt.Errorf("list active rootIDs by owner phone: rootType is invalid")
	}
	if ownerPhone == "" {
		return nil, fmt.Errorf("list active rootIDs by owner phone: ownerPhone is required")
	}
	rows, err := r.FindMany(ctx, activeRootScopeRelationFilter(bson.M{
		"root_type":   rootType,
		"owner_phone": ownerPhone,
	}))
	if err != nil {
		return nil, fmt.Errorf("list active rootIDs by owner phone: %w", err)
	}
	ids := make([]bson.ObjectID, 0, len(rows))
	seen := make(map[bson.ObjectID]struct{}, len(rows))
	for _, row := range rows {
		if row.RootID.IsZero() {
			continue
		}
		if _, ok := seen[row.RootID]; ok {
			continue
		}
		seen[row.RootID] = struct{}{}
		ids = append(ids, row.RootID)
	}
	return ids, nil
}

func (r *RootScopeRepository) CanAccessRoot(ctx context.Context, rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, ownerPhone string) (bool, error) {
	filter, err := activeRootScopeAccessFilter(rootType, rootID, ownerPhone)
	if err != nil {
		return false, fmt.Errorf("can access hpd root scope: %w", err)
	}
	total, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("can access hpd root scope: %w", err)
	}
	return total > 0, nil
}

func normalizeRootScopeRelation(entity *hpdmodel.HpdRootScopeRelation) {
	if entity == nil {
		return
	}
	entity.OwnerPhone = strings.TrimSpace(entity.OwnerPhone)
	if entity.RelationStatus == hpdmodel.HpdRelationStatusUnspecified {
		entity.RelationStatus = hpdmodel.HpdRelationStatusActive
	}
}
