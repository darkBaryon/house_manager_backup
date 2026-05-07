package repository

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// UserAuthRepository 用户认证绑定仓库。
type UserAuthRepository struct {
	*Repository[model.UserAuth]
}

func NewUserAuthRepository(client *dbmongo.Client) *UserAuthRepository {
	return &UserAuthRepository{
		Repository: NewRepository[model.UserAuth](client.Collection(model.CollectionUserAuth)),
	}
}

func (r *UserAuthRepository) Create(ctx context.Context, auth *model.UserAuth) error {
	if auth == nil {
		return fmt.Errorf("create user auth: auth is nil")
	}
	if auth.UserID.IsZero() {
		return fmt.Errorf("create user auth: userID is required")
	}
	if auth.AuthProvider == "" {
		return fmt.Errorf("create user auth: authProvider is required")
	}
	if auth.OpenID == "" {
		return fmt.Errorf("create user auth: openID is required")
	}
	return r.Insert(ctx, auth)
}

func (r *UserAuthRepository) TouchLastLogin(ctx context.Context, authID bson.ObjectID, lastLoginAt int64, lastLoginIP string) error {
	if authID.IsZero() {
		return fmt.Errorf("touch user auth last login: authID is required")
	}
	if lastLoginAt <= 0 {
		return fmt.Errorf("touch user auth last login: lastLoginAt is required")
	}
	return r.UpdateFieldsById(ctx, authID, bson.M{
		"last_login_at": lastLoginAt,
		"last_login_ip": lastLoginIP,
	})
}

func (r *UserAuthRepository) FindByOpenID(ctx context.Context, authProvider, openID string) (*model.UserAuth, error) {
	if authProvider == "" || openID == "" {
		return nil, fmt.Errorf("find user auth by openid: authProvider and openID are required")
	}
	return r.FindOne(ctx, bson.M{
		"auth_provider": authProvider,
		"openid":        openID,
		"status":        model.StatusActive,
	})
}

func (r *UserAuthRepository) FindByUnionID(ctx context.Context, authProvider, unionID string) (*model.UserAuth, error) {
	if authProvider == "" || unionID == "" {
		return nil, fmt.Errorf("find user auth by unionid: authProvider and unionID are required")
	}
	return r.FindOne(ctx, bson.M{
		"auth_provider": authProvider,
		"unionid":       unionID,
		"status":        model.StatusActive,
	})
}

func (r *UserAuthRepository) FindByUserID(ctx context.Context, userID bson.ObjectID) (*model.UserAuth, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("find user auth by userID: userID is required")
	}
	return r.FindOne(ctx, bson.M{
		"user_id": userID,
		"status":  model.StatusActive,
	})
}
