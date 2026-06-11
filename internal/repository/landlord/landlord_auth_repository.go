package landlord

import (
	"context"
	"fmt"
	"strings"
	"time"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type LandlordAuthRepository struct {
	*common.Repository[authmodel.LandlordAuth]
}

const (
	landlordAuthFieldLandlordID common.Field = "landlord_id"
	landlordAuthFieldAuthType   common.Field = "auth_type"
	landlordAuthFieldStatus     common.Field = "status"
)

func NewLandlordAuthRepository(client *dbmongo.Client) *LandlordAuthRepository {
	return &LandlordAuthRepository{
		Repository: common.NewRepository[authmodel.LandlordAuth](client.Collection(authmodel.CollectionLandlordAuth)),
	}
}

func (r *LandlordAuthRepository) Create(ctx context.Context, auth *authmodel.LandlordAuth) error {
	normalizeLandlordAuth(auth)
	if err := auth.ValidateForCreate(); err != nil {
		return fmt.Errorf("create landlord auth: %w", err)
	}
	return r.Insert(ctx, auth)
}

func (r *LandlordAuthRepository) FindActivePasswordByLandlordID(ctx context.Context, landlordID bson.ObjectID) (*authmodel.LandlordAuth, error) {
	if landlordID.IsZero() {
		return nil, fmt.Errorf("find landlord password auth: landlordID is required")
	}
	auth, err := r.FindOne(ctx, bson.M{
		"landlord_id": landlordID,
		"auth_type":   authmodel.PasswordAuthTypePassword,
		"status":      commonmodel.StatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("find landlord password auth: %w", err)
	}
	return auth, nil
}

func (r *LandlordAuthRepository) FindActivePasswordByLandlordIDs(ctx context.Context, landlordIDs []bson.ObjectID) ([]authmodel.LandlordAuth, error) {
	objectIDs := compactObjectIDs(landlordIDs)
	if len(objectIDs) == 0 {
		return []authmodel.LandlordAuth{}, nil
	}
	items, err := r.FindMany(ctx, bson.M{
		"landlord_id": bson.M{"$in": objectIDs},
		"auth_type":   authmodel.PasswordAuthTypePassword,
		"status":      commonmodel.StatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("find landlord password auth by landlord ids: %w", err)
	}
	return items, nil
}

func (r *LandlordAuthRepository) TouchLastLogin(ctx context.Context, authID bson.ObjectID, loginIP string) error {
	if authID.IsZero() {
		return fmt.Errorf("touch landlord auth last login: authID is required")
	}
	return r.UpdateFieldsByID(ctx, authID, bson.M{
		"last_login_at": time.Now().Unix(),
		"last_login_ip": strings.TrimSpace(loginIP),
	})
}

func (r *LandlordAuthRepository) RollbackCreateByLandlordID(ctx context.Context, landlordID bson.ObjectID) error {
	if landlordID.IsZero() {
		return fmt.Errorf("rollback landlord auth create: landlordID is required")
	}
	if err := r.DeleteAllBy(ctx, common.Eq(landlordAuthFieldLandlordID, landlordID)); err != nil {
		return fmt.Errorf("rollback landlord auth create: %w", err)
	}
	return nil
}

func normalizeLandlordAuth(auth *authmodel.LandlordAuth) {
	if auth == nil {
		return
	}
	if auth.AuthType == "" {
		auth.AuthType = authmodel.PasswordAuthTypePassword
	}
	auth.PasswordHash = strings.TrimSpace(auth.PasswordHash)
	auth.LastLoginIP = strings.TrimSpace(auth.LastLoginIP)
}

func compactObjectIDs(ids []bson.ObjectID) []bson.ObjectID {
	if len(ids) == 0 {
		return nil
	}
	result := make([]bson.ObjectID, 0, len(ids))
	seen := make(map[bson.ObjectID]struct{}, len(ids))
	for _, id := range ids {
		if id.IsZero() {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
