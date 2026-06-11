package history

import (
	"context"
	"fmt"
	commonmodel "house-manager/internal/model/common"
	useractivitymodel "house-manager/internal/model/useractivity"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Repository struct {
	*common.Repository[useractivitymodel.History]
}

func NewRepository(client *dbmongo.Client) *Repository {
	return &Repository{
		Repository: common.NewRepository[useractivitymodel.History](client.Collection(useractivitymodel.CollectionHistory)),
	}
}

func (r *Repository) Create(ctx context.Context, entity *useractivitymodel.History) error {
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create history: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *Repository) Upsert(ctx context.Context, entity *useractivitymodel.History) error {
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("upsert history: %w", err)
	}
	now := time.Now().Unix()
	update := common.NewUpdateDoc().
		Set(fieldUserID, entity.UserID).
		Set(fieldListingID, entity.ListingID).
		Set(fieldSource, entity.Source).
		Set(fieldViewedAt, entity.ViewedAt).
		Set(fieldStatus, commonmodel.StatusActive).
		Set(fieldUpdatedAt, now).
		Inc(fieldVersion, 1).
		SetOnInsert(fieldCreatedAt, now)
	if _, err := r.UpsertOneBy(ctx, userListingFilter(entity.UserID, entity.ListingID), update); err != nil {
		return fmt.Errorf("upsert history: %w", err)
	}
	return nil
}

func (r *Repository) List(ctx context.Context, userID bson.ObjectID, skip, limit int64) ([]useractivitymodel.History, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("list history: userID is required")
	}
	opts := []common.QueryOption{common.SortBy(fieldViewedAt, common.SortDesc)}
	if skip > 0 {
		opts = append(opts, common.Skip(skip))
	}
	if limit > 0 {
		opts = append(opts, common.Limit(limit))
	}
	return r.FindManyBy(ctx, common.And(
		common.Eq(fieldUserID, userID),
		common.Active(),
	), opts...)
}

func (r *Repository) Count(ctx context.Context, userID bson.ObjectID) (int64, error) {
	if userID.IsZero() {
		return 0, fmt.Errorf("count history: userID is required")
	}
	total, err := r.CountBy(ctx, common.And(
		common.Eq(fieldUserID, userID),
		common.Active(),
	))
	if err != nil {
		return 0, fmt.Errorf("count history: %w", err)
	}
	return total, nil
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex("user_id_1_listing_id_1",
			common.IndexKey(fieldUserID, common.SortAsc),
			common.IndexKey(fieldListingID, common.SortAsc),
		).WithUnique(),
		common.NewIndex("user_id_1_viewed_at_-1",
			common.IndexKey(fieldUserID, common.SortAsc),
			common.IndexKey(fieldViewedAt, common.SortDesc),
		),
	); err != nil {
		return fmt.Errorf("ensure history indexes: %w", err)
	}
	return nil
}

func userListingFilter(userID, listingID bson.ObjectID) common.Filter {
	return common.And(
		common.Eq(fieldUserID, userID),
		common.Eq(fieldListingID, listingID),
	)
}
