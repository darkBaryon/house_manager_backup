package auth

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// UserRepository 用户主档仓库。
type UserRepository struct {
	*common.Repository[model.User]
}

func NewUserRepository(client *dbmongo.Client) *UserRepository {
	return &UserRepository{
		Repository: common.NewRepository[model.User](client.Collection(model.CollectionUser)),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
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

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	if phone == "" {
		return nil, fmt.Errorf("find user by phone: phone is required")
	}
	return r.FindOne(ctx, bson.M{
		"phone":  phone,
		"status": model.StatusActive,
	})
}
