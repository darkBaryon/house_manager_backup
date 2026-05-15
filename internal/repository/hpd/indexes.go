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

func (r *EntrustRelationRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "listing_id", Value: 1}},
			Options: options.Index().
				SetName("listing_id_1_active_unique").
				SetUnique(true).
				SetPartialFilterExpression(bson.M{
					"status":          commonmodel.StatusActive,
					"relation_status": hpdmodel.HpdRelationStatusActive,
				}),
		},
		{
			Keys: bson.D{
				{Key: "service_staff_id", Value: 1},
				{Key: "relation_status", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("service_staff_id_1_relation_status_1_status_1"),
		},
		{
			Keys: bson.D{
				{Key: "maintainer_staff_id", Value: 1},
				{Key: "relation_status", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("maintainer_staff_id_1_relation_status_1_status_1"),
		},
		{
			Keys: bson.D{
				{Key: "owner_phone", Value: 1},
				{Key: "relation_status", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("owner_phone_1_relation_status_1_status_1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure hpd entrust relation indexes: %w", err)
	}
	return nil
}
