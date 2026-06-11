package landlord

import (
	"context"
	"fmt"
	"strings"

	authmodel "house-manager/internal/model/auth"
	"house-manager/internal/repository/common"
	"house-manager/internal/repository/credentialcore"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type LandlordAuthRepository struct {
	*common.Repository[authmodel.LandlordAuth]
	core *credentialcore.Core[authmodel.LandlordAuth]
}

const (
	landlordAuthFieldLandlordID common.Field = "landlord_id"
	landlordAuthFieldAuthType   common.Field = "auth_type"
	landlordAuthFieldStatus     common.Field = "status"
)

func NewLandlordAuthRepository(client *dbmongo.Client) *LandlordAuthRepository {
	repo := common.NewRepository[authmodel.LandlordAuth](client.Collection(authmodel.CollectionLandlordAuth))
	return &LandlordAuthRepository{
		Repository: repo,
		core: credentialcore.NewCore(repo, credentialcore.Config[authmodel.LandlordAuth]{
			OwnerField: landlordAuthFieldLandlordID,
			Normalize:  normalizeLandlordAuth,
			Validate:   (*authmodel.LandlordAuth).ValidateForCreate,
		}),
	}
}

func (r *LandlordAuthRepository) Create(ctx context.Context, auth *authmodel.LandlordAuth) error {
	if err := r.core.Create(ctx, auth); err != nil {
		return fmt.Errorf("create landlord auth: %w", err)
	}
	return nil
}

func (r *LandlordAuthRepository) FindActivePasswordByLandlordID(ctx context.Context, landlordID bson.ObjectID) (*authmodel.LandlordAuth, error) {
	if landlordID.IsZero() {
		return nil, fmt.Errorf("find landlord password auth: landlordID is required")
	}
	auth, err := r.core.FindActivePasswordByOwner(ctx, landlordID)
	if err != nil {
		return nil, fmt.Errorf("find landlord password auth: %w", err)
	}
	return auth, nil
}

func (r *LandlordAuthRepository) FindActivePasswordByLandlordIDs(ctx context.Context, landlordIDs []bson.ObjectID) ([]authmodel.LandlordAuth, error) {
	items, err := r.core.FindActivePasswordByOwners(ctx, landlordIDs)
	if err != nil {
		return nil, fmt.Errorf("find landlord password auth by landlord ids: %w", err)
	}
	return items, nil
}

func (r *LandlordAuthRepository) TouchLastLogin(ctx context.Context, authID bson.ObjectID, loginIP string) error {
	if authID.IsZero() {
		return fmt.Errorf("touch landlord auth last login: authID is required")
	}
	if err := r.core.TouchLastLogin(ctx, authID, strings.TrimSpace(loginIP)); err != nil {
		return fmt.Errorf("touch landlord auth last login: %w", err)
	}
	return nil
}

func (r *LandlordAuthRepository) RollbackCreateByLandlordID(ctx context.Context, landlordID bson.ObjectID) error {
	if landlordID.IsZero() {
		return fmt.Errorf("rollback landlord auth create: landlordID is required")
	}
	if err := r.core.RollbackCreateByOwner(ctx, landlordID); err != nil {
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
