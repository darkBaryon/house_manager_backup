package publishauth

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *OwnerUserRepository) EnsureIndexes(ctx context.Context) error {
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
		return fmt.Errorf("ensure publish owner user indexes: %w", err)
	}
	return nil
}
