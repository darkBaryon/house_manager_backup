package miniappauth

import (
	"context"
	"fmt"
	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// UserProfileExtRepository 用户偏好扩展仓库。
type UserProfileExtRepository struct {
	*common.Repository[authmodel.UserProfileExt]
}

func NewUserProfileExtRepository(client *dbmongo.Client) *UserProfileExtRepository {
	return &UserProfileExtRepository{
		Repository: common.NewRepository[authmodel.UserProfileExt](client.Collection(authmodel.CollectionUserProfileExt)),
	}
}

func (r *UserProfileExtRepository) Create(ctx context.Context, profile *authmodel.UserProfileExt) error {
	if err := profile.ValidateForCreate(); err != nil {
		return fmt.Errorf("create user profile ext: %w", err)
	}
	return r.Insert(ctx, profile)
}

func (r *UserProfileExtRepository) UpsertByUserID(ctx context.Context, userID bson.ObjectID, fields bson.M) (matched bool, err error) {
	if userID.IsZero() {
		return false, fmt.Errorf("upsert user profile ext by userID: userID is required")
	}
	if fields == nil {
		return false, fmt.Errorf("upsert user profile ext by userID: fields is nil")
	}
	upsertFields := make(bson.M, len(fields)+1)
	for k, v := range fields {
		upsertFields[k] = v
	}
	upsertFields["user_id"] = userID
	return r.UpsertFields(ctx, bson.M{
		"user_id": userID,
		"status":  bson.M{"$ne": commonmodel.StatusDeleted},
	}, upsertFields)
}

func (r *UserProfileExtRepository) FindByUserID(ctx context.Context, userID bson.ObjectID) (*authmodel.UserProfileExt, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("find user profile ext by userID: userID is required")
	}
	return r.FindOne(ctx, bson.M{
		"user_id": userID,
		"status":  commonmodel.StatusActive,
	})
}
