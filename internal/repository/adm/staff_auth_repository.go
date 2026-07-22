package adm

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

type StaffAuthRepository struct {
	*common.Repository[authmodel.AdmStaffAuth]
	core *credentialcore.Core[authmodel.AdmStaffAuth]
}

const (
	staffAuthFieldStaffID  common.Field = "staff_id"
	staffAuthFieldAuthType common.Field = "auth_type"
	staffAuthFieldStatus   common.Field = "status"
)

func NewStaffAuthRepository(client *dbmongo.Client) *StaffAuthRepository {
	repo := common.NewRepository[authmodel.AdmStaffAuth](client.Collection(authmodel.CollectionAdmStaffAuth))
	return &StaffAuthRepository{
		Repository: repo,
		core: credentialcore.NewCore(repo, credentialcore.Config[authmodel.AdmStaffAuth]{
			OwnerField: staffAuthFieldStaffID,
			Normalize:  normalizeStaffAuth,
			Validate:   (*authmodel.AdmStaffAuth).ValidateForCreate,
		}),
	}
}

func (r *StaffAuthRepository) CreatePasswordAuth(ctx context.Context, authRecord *authmodel.AdmStaffAuth) error {
	if err := r.core.Create(ctx, authRecord); err != nil {
		return fmt.Errorf("create staff password auth: %w", err)
	}
	return nil
}

func (r *StaffAuthRepository) FindActivePasswordByStaffID(ctx context.Context, staffID bson.ObjectID) (*authmodel.AdmStaffAuth, error) {
	if staffID.IsZero() {
		return nil, fmt.Errorf("find staff password auth: staffID is required")
	}
	authRecord, err := r.core.FindActivePasswordByOwner(ctx, staffID)
	if err != nil {
		return nil, fmt.Errorf("find staff password auth: %w", err)
	}
	return authRecord, nil
}

func (r *StaffAuthRepository) FindActivePasswordByStaffIDs(ctx context.Context, staffIDs []bson.ObjectID) ([]authmodel.AdmStaffAuth, error) {
	items, err := r.core.FindActivePasswordByOwners(ctx, staffIDs)
	if err != nil {
		return nil, fmt.Errorf("find staff password auth by staff ids: %w", err)
	}
	return items, nil
}

func (r *StaffAuthRepository) TouchLastLogin(ctx context.Context, authID bson.ObjectID, loginIP string) error {
	if authID.IsZero() {
		return fmt.Errorf("touch staff auth last login: authID is required")
	}
	if err := r.core.TouchLastLogin(ctx, authID, strings.TrimSpace(loginIP)); err != nil {
		return fmt.Errorf("touch staff auth last login: %w", err)
	}
	return nil
}

func (r *StaffAuthRepository) RollbackCreateByStaffID(ctx context.Context, staffID bson.ObjectID) error {
	if staffID.IsZero() {
		return fmt.Errorf("rollback staff auth create: staffID is required")
	}
	if err := r.core.RollbackCreateByOwner(ctx, staffID); err != nil {
		return fmt.Errorf("rollback staff auth create: %w", err)
	}
	return nil
}

func normalizeStaffAuth(authRecord *authmodel.AdmStaffAuth) {
	if authRecord == nil {
		return
	}
	authRecord.PasswordHash = strings.TrimSpace(authRecord.PasswordHash)
	authRecord.LastLoginIP = strings.TrimSpace(authRecord.LastLoginIP)
}
