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

// UserRepository 用户主档仓库。
type UserRepository struct {
	*common.Repository[authmodel.User]
}

func NewUserRepository(client *dbmongo.Client) *UserRepository {
	return &UserRepository{
		Repository: common.NewRepository[authmodel.User](client.Collection(authmodel.CollectionUser)),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *authmodel.User) error {
	if err := user.ValidateForCreate(); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return r.Insert(ctx, user)
}

func (r *UserRepository) TouchLastActive(ctx context.Context, userID bson.ObjectID, lastActiveAt int64) error {
	if userID.IsZero() {
		return fmt.Errorf("touch user last active: userID is required")
	}
	if lastActiveAt <= 0 {
		return fmt.Errorf("touch user last active: lastActiveAt is required")
	}
	return r.UpdateFieldsByID(ctx, userID, bson.M{
		"last_active_at": lastActiveAt,
	})
}

func (r *UserRepository) UpdateProfileFields(ctx context.Context, userID bson.ObjectID, fields bson.M) error {
	if userID.IsZero() {
		return fmt.Errorf("update user profile fields: userID is required")
	}
	if fields == nil {
		return fmt.Errorf("update user profile fields: fields is nil")
	}
	return r.UpdateFieldsByID(ctx, userID, fields)
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*authmodel.User, error) {
	if phone == "" {
		return nil, fmt.Errorf("find user by phone: phone is required")
	}
	return r.FindOne(ctx, bson.M{
		"phone":  phone,
		"status": commonmodel.StatusActive,
	})
}
