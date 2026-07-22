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
	update := common.NewUpdateDoc().
		Set(fieldUserID, userID).
		Set(fieldListingID, listingID).
		Set(fieldStatus, commonmodel.StatusActive).
		Set(fieldUpdatedAt, now).
		Inc(fieldVersion, 1).
		SetOnInsert(fieldCreatedAt, now)
	if _, err := r.UpsertOneBy(ctx, userListingFilter(userID, listingID), update); err != nil {
		return fmt.Errorf("upsert favorite: %w", err)
	}
	return nil
}

func (r *Repository) SoftRemove(ctx context.Context, userID, listingID bson.ObjectID) error {
	if userID.IsZero() || listingID.IsZero() {
		return fmt.Errorf("soft remove favorite: userID and listingID are required")
	}
	now := time.Now().Unix()
	update := common.NewUpdateDoc().
		Set(fieldStatus, commonmodel.StatusDeleted).
		Set(fieldUpdatedAt, now).
		Inc(fieldVersion, 1)
	if _, err := r.UpdateOneBy(ctx, userListingFilter(userID, listingID), update); err != nil {
		return fmt.Errorf("soft remove favorite: %w", err)
	}
	return nil
}

func (r *Repository) Exists(ctx context.Context, userID, listingID bson.ObjectID) (bool, error) {
	if userID.IsZero() || listingID.IsZero() {
		return false, fmt.Errorf("favorite exists: userID and listingID are required")
	}
	exists, err := r.ExistsBy(ctx, common.And(
		common.Eq(fieldUserID, userID),
		common.Eq(fieldListingID, listingID),
		common.Active(),
	))
	if err != nil {
		return false, fmt.Errorf("favorite exists: %w", err)
	}
	return exists, nil
}

func (r *Repository) List(ctx context.Context, userID bson.ObjectID, skip, limit int64) ([]useractivitymodel.Favorite, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("list favorites: userID is required")
	}
	opts := []common.QueryOption{common.SortBy(fieldUpdatedAt, common.SortDesc)}
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
		return 0, fmt.Errorf("count favorites: userID is required")
	}
	total, err := r.CountBy(ctx, common.And(
		common.Eq(fieldUserID, userID),
		common.Active(),
	))
	if err != nil {
		return 0, fmt.Errorf("count favorites: %w", err)
	}
	return total, nil
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex("user_id_1_listing_id_1",
			common.IndexKey(fieldUserID, common.SortAsc),
			common.IndexKey(fieldListingID, common.SortAsc),
		).WithUnique(),
		common.NewIndex("user_id_1_status_1_updated_at_-1",
			common.IndexKey(fieldUserID, common.SortAsc),
			common.IndexKey(fieldStatus, common.SortAsc),
			common.IndexKey(fieldUpdatedAt, common.SortDesc),
		),
	); err != nil {
		return fmt.Errorf("ensure favorite indexes: %w", err)
	}
	return nil
}

func userListingFilter(userID, listingID bson.ObjectID) common.Filter {
	return common.And(
		common.Eq(fieldUserID, userID),
		common.Eq(fieldListingID, listingID),
	)
}
