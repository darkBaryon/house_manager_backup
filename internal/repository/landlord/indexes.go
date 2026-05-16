package landlord

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *LandlordRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "phone", Value: 1}},
			Options: options.Index().SetName("phone_1").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("status_1_updated_at_-1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure publish landlord indexes: %w", err)
	}
	return nil
}

func (r *LandlordAuthRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "landlord_id", Value: 1}},
			Options: options.Index().SetName("landlord_id_1").SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "auth_type", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("auth_type_1_status_1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure publish landlord auth indexes: %w", err)
	}
	return nil
}
