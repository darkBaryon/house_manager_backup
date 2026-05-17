package hpd

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	commonmodel "house-manager/internal/model/common"
	hpdmodel "house-manager/internal/model/hpd"
)

func (r *RootScopeRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "root_type", Value: 1},
				{Key: "root_id", Value: 1},
			},
			Options: options.Index().
				SetName("root_type_1_root_id_1_active_unique").
				SetUnique(true).
				SetPartialFilterExpression(bson.M{
					"status":          commonmodel.StatusActive,
					"relation_status": hpdmodel.HpdRelationStatusActive,
				}),
		},
		{
			Keys: bson.D{
				{Key: "root_type", Value: 1},
				{Key: "owner_landlord_id", Value: 1},
				{Key: "relation_status", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("root_type_1_owner_landlord_id_1_relation_status_1_status_1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure hpd root scope relation indexes: %w", err)
	}
	return nil
}

func (r *PublisherListingRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "listing_id", Value: 1}},
			Options: options.Index().
				SetName("listing_id_1").
				SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "source_type", Value: 1},
				{Key: "source_id", Value: 1},
			},
			Options: options.Index().SetName("source_type_1_source_id_1"),
		},
		{
			Keys: bson.D{
				{Key: "owner_landlord_id", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index().SetName("owner_landlord_id_1_updated_at_-1"),
		},
		{
			Keys: bson.D{
				{Key: "root_type", Value: 1},
				{Key: "root_id", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index().SetName("root_type_1_root_id_1_updated_at_-1"),
		},
		{
			Keys: bson.D{
				{Key: "project_id", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index().SetName("project_id_1_updated_at_-1"),
		},
		{
			Keys: bson.D{
				{Key: "building_id", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index().SetName("building_id_1_updated_at_-1"),
		},
		{
			Keys: bson.D{
				{Key: "decentralized_id", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index().SetName("decentralized_id_1_updated_at_-1"),
		},
		{
			Keys: bson.D{
				{Key: "room_status", Value: 1},
				{Key: "listing_status", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index().SetName("room_status_1_listing_status_1_updated_at_-1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure hpd publisher listing indexes: %w", err)
	}
	return nil
}

func (r *AdminListingRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "listing_id", Value: 1}},
			Options: options.Index().
				SetName("listing_id_1").
				SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "source_type", Value: 1},
				{Key: "source_id", Value: 1},
			},
			Options: options.Index().SetName("source_type_1_source_id_1"),
		},
		{
			Keys: bson.D{
				{Key: "owner_landlord_id", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index().SetName("owner_landlord_id_1_updated_at_-1"),
		},
		{
			Keys: bson.D{
				{Key: "asset_mode", Value: 1},
				{Key: "city", Value: 1},
				{Key: "district", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index().SetName("asset_mode_1_city_1_district_1_updated_at_-1"),
		},
		{
			Keys: bson.D{
				{Key: "listing_status", Value: 1},
				{Key: "audit_status", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index().SetName("listing_status_1_audit_status_1_updated_at_-1"),
		},
		{
			Keys: bson.D{
				{Key: "room_status", Value: 1},
				{Key: "is_online", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index().SetName("room_status_1_is_online_1_updated_at_-1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure hpd admin listing indexes: %w", err)
	}
	return nil
}
