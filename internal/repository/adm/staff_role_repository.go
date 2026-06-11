package adm

import (
	"context"
	"fmt"

	authmodel "house-manager/internal/model/auth"
	"house-manager/internal/repository/common"
	"house-manager/internal/repository/relationcore"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type StaffRoleRepository struct {
	*common.Repository[authmodel.AdmStaffRole]
	core *relationcore.Core[authmodel.AdmStaffRole]
}

const (
	staffRoleFieldStaffID    common.Field = "staff_id"
	staffRoleFieldRoleID     common.Field = "role_id"
	staffRoleFieldStatus     common.Field = "status"
	staffRoleFieldUpdatedAt  common.Field = "updated_at"
	staffRoleFieldVersion    common.Field = "version"
	staffRoleFieldAssignedBy common.Field = "assigned_by"
	staffRoleFieldAssignedAt common.Field = "assigned_at"
)

func NewStaffRoleRepository(client *dbmongo.Client) *StaffRoleRepository {
	repo := common.NewRepository[authmodel.AdmStaffRole](client.Collection(authmodel.CollectionAdmStaffRole))
	return &StaffRoleRepository{
		Repository: repo,
		core: relationcore.NewCore(repo, relationcore.Config[authmodel.AdmStaffRole]{
			LeftField:       staffRoleFieldStaffID,
			RightField:      staffRoleFieldRoleID,
			AssignedByField: staffRoleFieldAssignedBy,
			AssignedAtField: staffRoleFieldAssignedAt,
			Validate:        (*authmodel.AdmStaffRole).ValidateForCreate,
		}),
	}
}

func (r *StaffRoleRepository) CreateMany(ctx context.Context, staffID bson.ObjectID, roleIDs []bson.ObjectID, assignedBy bson.ObjectID) error {
	if staffID.IsZero() {
		return fmt.Errorf("create staff roles: staffID is required")
	}
	if err := r.core.CreateMany(ctx, staffID, roleIDs, assignedBy); err != nil {
		return fmt.Errorf("create staff roles: %w", err)
	}
	return nil
}

func (r *StaffRoleRepository) ListActiveByStaffID(ctx context.Context, staffID bson.ObjectID) ([]authmodel.AdmStaffRole, error) {
	if staffID.IsZero() {
		return nil, fmt.Errorf("list staff roles: staffID is required")
	}
	items, err := r.core.ListActiveByLeft(ctx, staffID)
	if err != nil {
		return nil, fmt.Errorf("list staff roles: %w", err)
	}
	return items, nil
}

func (r *StaffRoleRepository) ListActiveByStaffIDs(ctx context.Context, staffIDs []bson.ObjectID) ([]authmodel.AdmStaffRole, error) {
	items, err := r.core.ListActiveByLefts(ctx, staffIDs)
	if err != nil {
		return nil, fmt.Errorf("list staff roles by staff ids: %w", err)
	}
	return items, nil
}

func (r *StaffRoleRepository) ListActiveByRoleID(ctx context.Context, roleID bson.ObjectID) ([]authmodel.AdmStaffRole, error) {
	if roleID.IsZero() {
		return nil, fmt.Errorf("list staff roles by role: roleID is required")
	}
	items, err := r.core.ListActiveByRight(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("list staff roles by role: %w", err)
	}
	return items, nil
}

func (r *StaffRoleRepository) ReplaceByStaffID(ctx context.Context, staffID bson.ObjectID, roleIDs []bson.ObjectID, assignedBy bson.ObjectID) error {
	if staffID.IsZero() {
		return fmt.Errorf("replace staff roles: staffID is required")
	}
	if err := r.core.ReplaceByLeft(ctx, staffID, roleIDs, assignedBy); err != nil {
		return fmt.Errorf("replace staff roles: %w", err)
	}
	return nil
}

func (r *StaffRoleRepository) RollbackCreateByStaffID(ctx context.Context, staffID bson.ObjectID) error {
	if staffID.IsZero() {
		return fmt.Errorf("rollback staff roles create: staffID is required")
	}
	if err := r.core.RollbackByLeft(ctx, staffID); err != nil {
		return fmt.Errorf("rollback staff roles create: %w", err)
	}
	return nil
}
