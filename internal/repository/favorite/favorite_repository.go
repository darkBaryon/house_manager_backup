package favorite

import (
	"context"
	"fmt"
	commonmodel "house-manager/internal/model/common"
	useractivitymodel "house-manager/internal/model/useractivity"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	*common.Repository[useractivitymodel.Favorite]
}

func NewRepository(client *dbmongo.Client) *Repository {
	return &Repository{
		Repository: common.NewRepository[useractivitymodel.Favorite](client.Collection(useractivitymodel.CollectionFavorite)),
	}
}

func (r *Repository) Create(ctx context.Context, entity *useractivitymodel.Favorite) error {
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create favorite: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *Repository) Upsert(ctx context.Context, userID, listingID bson.ObjectID) error {
	if userID.IsZero() || listingID.IsZero() {
		return fmt.Errorf("upsert favorite: userID and listingID are required")
	}
	now := time.Now().Unix()
	update := bson.M{
		"$set": bson.M{
			"user_id":    userID,
			"listing_id": listingID,
			"status":     commonmodel.StatusActive,
			"updated_at": now,
		},
		"$inc": bson.M{"version": 1},
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}
	if _, err := r.Collection.UpdateOne(ctx, userListingFilter(userID, listingID), update, options.UpdateOne().SetUpsert(true)); err != nil {
		return fmt.Errorf("upsert favorite: %w", err)
	}
	return nil
}

func (r *Repository) SoftRemove(ctx context.Context, userID, listingID bson.ObjectID) error {
	if userID.IsZero() || listingID.IsZero() {
		return fmt.Errorf("soft remove favorite: userID and listingID are required")
	}
	now := time.Now().Unix()
	update := bson.M{
		"$set": bson.M{
			"status":     commonmodel.StatusDeleted,
			"updated_at": now,
		},
		"$inc": bson.M{"version": 1},
	}
	if _, err := r.Collection.UpdateOne(ctx, userListingFilter(userID, listingID), update); err != nil {
		return fmt.Errorf("soft remove favorite: %w", err)
	}
	return nil
}

func (r *Repository) Exists(ctx context.Context, userID, listingID bson.ObjectID) (bool, error) {
	if userID.IsZero() || listingID.IsZero() {
		return false, fmt.Errorf("favorite exists: userID and listingID are required")
	}
	total, err := r.Collection.CountDocuments(ctx, bson.M{
		"user_id":    userID,
		"listing_id": listingID,
		"status":     commonmodel.StatusActive,
	})
	if err != nil {
		return false, fmt.Errorf("favorite exists: %w", err)
	}
	return total > 0, nil
}

func (r *Repository) List(ctx context.Context, userID bson.ObjectID, skip, limit int64) ([]useractivitymodel.Favorite, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("list favorites: userID is required")
	}
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}
	return r.FindMany(ctx, bson.M{"user_id": userID, "status": commonmodel.StatusActive}, opts)
}

func (r *Repository) Count(ctx context.Context, userID bson.ObjectID) (int64, error) {
	if userID.IsZero() {
		return 0, fmt.Errorf("count favorites: userID is required")
	}
	total, err := r.Collection.CountDocuments(ctx, bson.M{"user_id": userID, "status": commonmodel.StatusActive})
	if err != nil {
		return 0, fmt.Errorf("count favorites: %w", err)
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
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "status", Value: 1}, {Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("user_id_1_status_1_updated_at_-1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure favorite indexes: %w", err)
	}
	return nil
}

func userListingFilter(userID, listingID bson.ObjectID) bson.M {
	return bson.M{"user_id": userID, "listing_id": listingID}
}
