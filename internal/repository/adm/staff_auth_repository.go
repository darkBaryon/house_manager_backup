package adm

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

type StaffAuthRepository struct {
	*common.Repository[authmodel.AdmStaffAuth]
}

func NewStaffAuthRepository(client *dbmongo.Client) *StaffAuthRepository {
	return &StaffAuthRepository{
		Repository: common.NewRepository[authmodel.AdmStaffAuth](client.Collection(authmodel.CollectionAdmStaffAuth)),
	}
}

func (r *StaffAuthRepository) CreatePasswordAuth(ctx context.Context, authRecord *authmodel.AdmStaffAuth) error {
	normalizeStaffAuth(authRecord)
	if err := authRecord.ValidateForCreate(); err != nil {
		return fmt.Errorf("create staff password auth: %w", err)
	}
	return r.Insert(ctx, authRecord)
}

func (r *StaffAuthRepository) FindActivePasswordByStaffID(ctx context.Context, staffID bson.ObjectID) (*authmodel.AdmStaffAuth, error) {
	if staffID.IsZero() {
		return nil, fmt.Errorf("find staff password auth: staffID is required")
	}
	authRecord, err := r.FindOne(ctx, bson.M{
		"staff_id":  staffID,
		"auth_type": authmodel.PasswordAuthTypePassword,
		"status":    commonmodel.StatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("find staff password auth: %w", err)
	}
	return authRecord, nil
}

func (r *StaffAuthRepository) FindActivePasswordByStaffIDs(ctx context.Context, staffIDs []bson.ObjectID) ([]authmodel.AdmStaffAuth, error) {
	objectIDs := compactObjectIDs(staffIDs)
	if len(objectIDs) == 0 {
		return []authmodel.AdmStaffAuth{}, nil
	}
	items, err := r.FindMany(ctx, bson.M{
		"staff_id":  bson.M{"$in": objectIDs},
		"auth_type": authmodel.PasswordAuthTypePassword,
		"status":    commonmodel.StatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("find staff password auth by staff ids: %w", err)
	}
	return items, nil
}

func (r *StaffAuthRepository) TouchLastLogin(ctx context.Context, authID bson.ObjectID, loginIP string) error {
	if authID.IsZero() {
		return fmt.Errorf("touch staff auth last login: authID is required")
	}
	return r.UpdateFieldsByID(ctx, authID, bson.M{
		"last_login_at": time.Now().Unix(),
		"last_login_ip": strings.TrimSpace(loginIP),
	})
}

func (r *StaffAuthRepository) RollbackCreateByStaffID(ctx context.Context, staffID bson.ObjectID) error {
	if staffID.IsZero() {
		return fmt.Errorf("rollback staff auth create: staffID is required")
	}
	if _, err := r.Collection.DeleteMany(ctx, bson.M{"staff_id": staffID}); err != nil {
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
