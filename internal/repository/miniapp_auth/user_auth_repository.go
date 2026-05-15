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

// UserAuthRepository 用户认证绑定仓库。
type UserAuthRepository struct {
	*common.Repository[authmodel.UserAuth]
}

func NewUserAuthRepository(client *dbmongo.Client) *UserAuthRepository {
	return &UserAuthRepository{
		Repository: common.NewRepository[authmodel.UserAuth](client.Collection(authmodel.CollectionUserAuth)),
	}
}

func (r *UserAuthRepository) Create(ctx context.Context, auth *authmodel.UserAuth) error {
	if err := auth.ValidateForCreate(); err != nil {
		return fmt.Errorf("create user auth: %w", err)
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
	return r.UpdateFieldsByID(ctx, authID, bson.M{
		"last_login_at": lastLoginAt,
		"last_login_ip": lastLoginIP,
	})
}

func (r *UserAuthRepository) UpdateWechatIdentity(ctx context.Context, authID bson.ObjectID, openID, unionID string, lastLoginAt int64, lastLoginIP string) error {
	if authID.IsZero() {
		return fmt.Errorf("update wechat identity: authID is required")
	}
	if openID == "" {
		return fmt.Errorf("update wechat identity: openID is required")
	}
	if lastLoginAt <= 0 {
		return fmt.Errorf("update wechat identity: lastLoginAt is required")
	}

	fields := bson.M{
		"openid":        openID,
		"last_login_at": lastLoginAt,
		"last_login_ip": lastLoginIP,
	}
	if unionID != "" {
		fields["unionid"] = unionID
	}
	return r.UpdateFieldsByID(ctx, authID, fields)
}

func (r *UserAuthRepository) FindByOpenID(ctx context.Context, authProvider authmodel.AuthProvider, openID string) (*authmodel.UserAuth, error) {
	if !authProvider.Valid() || openID == "" {
		return nil, fmt.Errorf("find user auth by openid: authProvider and openID are required")
	}
	return r.FindOne(ctx, bson.M{
		"auth_provider": authProvider,
		"openid":        openID,
		"status":        commonmodel.StatusActive,
	})
}

func (r *UserAuthRepository) FindByUnionID(ctx context.Context, authProvider authmodel.AuthProvider, unionID string) (*authmodel.UserAuth, error) {
	if !authProvider.Valid() || unionID == "" {
		return nil, fmt.Errorf("find user auth by unionid: authProvider and unionID are required")
	}
	return r.FindOne(ctx, bson.M{
		"auth_provider": authProvider,
		"unionid":       unionID,
		"status":        commonmodel.StatusActive,
	})
}

func (r *UserAuthRepository) FindByUserID(ctx context.Context, userID bson.ObjectID) (*authmodel.UserAuth, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("find user auth by userID: userID is required")
	}
	return r.FindOne(ctx, bson.M{
		"user_id": userID,
		"status":  commonmodel.StatusActive,
	})
}
