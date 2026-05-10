package history

import (
	"context"
	"fmt"
	"time"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	*common.Repository[model.History]
}

func NewRepository(client *dbmongo.Client) *Repository {
	return &Repository{
		Repository: common.NewRepository[model.History](client.Collection(model.CollectionHistory)),
	}
}

func (r *Repository) Create(ctx context.Context, entity *model.History) error {
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create history: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *Repository) Upsert(ctx context.Context, entity *model.History) error {
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("upsert history: %w", err)
	}
	now := time.Now().Unix()
	update := bson.M{
		"$set": bson.M{
			"user_id":    entity.UserID,
			"listing_id": entity.ListingID,
			"source":     entity.Source,
			"viewed_at":  entity.ViewedAt,
			"status":     model.StatusActive,
			"updated_at": now,
		},
		"$inc": bson.M{"version": 1},
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}
	if _, err := r.Collection.UpdateOne(ctx, userListingFilter(entity.UserID, entity.ListingID), update, options.UpdateOne().SetUpsert(true)); err != nil {
		return fmt.Errorf("upsert history: %w", err)
	}
	return nil
}

func (r *Repository) List(ctx context.Context, userID bson.ObjectID, skip, limit int64) ([]model.History, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("list history: userID is required")
	}
	opts := options.Find().SetSort(bson.D{{Key: "viewed_at", Value: -1}})
	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}
	return r.FindMany(ctx, bson.M{"user_id": userID, "status": model.StatusActive}, opts)
}

func (r *Repository) Count(ctx context.Context, userID bson.ObjectID) (int64, error) {
	if userID.IsZero() {
		return 0, fmt.Errorf("count history: userID is required")
	}
	total, err := r.Collection.CountDocuments(ctx, bson.M{"user_id": userID, "status": model.StatusActive})
	if err != nil {
		return 0, fmt.Errorf("count history: %w", err)
	}
	return total, nil
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "listing_id", Value: 1}},
			Options: options.Index().SetName("user_id_1_listing_id_1").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "viewed_at", Value: -1}},
			Options: options.Index().SetName("user_id_1_viewed_at_-1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure history indexes: %w", err)
	}
	return nil
}

func userListingFilter(userID, listingID bson.ObjectID) bson.M {
	return bson.M{"user_id": userID, "listing_id": listingID}
}
