package adm

import (
	"context"
	"fmt"
	"time"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type StaffRoleRepository struct {
	*common.Repository[authmodel.AdmStaffRole]
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
	return &StaffRoleRepository{
		Repository: common.NewRepository[authmodel.AdmStaffRole](client.Collection(authmodel.CollectionAdmStaffRole)),
	}
}

func (r *StaffRoleRepository) CreateMany(ctx context.Context, staffID bson.ObjectID, roleIDs []bson.ObjectID, assignedBy bson.ObjectID) error {
	if staffID.IsZero() {
		return fmt.Errorf("create staff roles: staffID is required")
	}
	roleIDs = compactObjectIDs(roleIDs)
	if len(roleIDs) == 0 {
		return nil
	}
	now := time.Now().Unix()
	for _, roleID := range roleIDs {
		item := &authmodel.AdmStaffRole{
			StaffID:    staffID,
			RoleID:     roleID,
			AssignedBy: assignedBy,
			AssignedAt: now,
		}
		if err := item.ValidateForCreate(); err != nil {
			return fmt.Errorf("create staff roles: %w", err)
		}
		if err := r.Insert(ctx, item); err != nil {
			return fmt.Errorf("create staff roles: %w", err)
		}
	}
	return nil
}

func (r *StaffRoleRepository) ListActiveByStaffID(ctx context.Context, staffID bson.ObjectID) ([]authmodel.AdmStaffRole, error) {
	if staffID.IsZero() {
		return nil, fmt.Errorf("list staff roles: staffID is required")
	}
	items, err := r.FindMany(ctx, bson.M{
		"staff_id": staffID,
		"status":   commonmodel.StatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("list staff roles: %w", err)
	}
	return items, nil
}

func (r *StaffRoleRepository) ListActiveByStaffIDs(ctx context.Context, staffIDs []bson.ObjectID) ([]authmodel.AdmStaffRole, error) {
	objectIDs := compactObjectIDs(staffIDs)
	if len(objectIDs) == 0 {
		return []authmodel.AdmStaffRole{}, nil
	}
	items, err := r.FindMany(ctx, bson.M{
		"staff_id": bson.M{"$in": objectIDs},
		"status":   commonmodel.StatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("list staff roles by staff ids: %w", err)
	}
	return items, nil
}

func (r *StaffRoleRepository) ListActiveByRoleID(ctx context.Context, roleID bson.ObjectID) ([]authmodel.AdmStaffRole, error) {
	if roleID.IsZero() {
		return nil, fmt.Errorf("list staff roles by role: roleID is required")
	}
	items, err := r.FindMany(ctx, bson.M{
		"role_id": roleID,
		"status":  commonmodel.StatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("list staff roles by role: %w", err)
	}
	return items, nil
}

func (r *StaffRoleRepository) ReplaceByStaffID(ctx context.Context, staffID bson.ObjectID, roleIDs []bson.ObjectID, assignedBy bson.ObjectID) error {
	if staffID.IsZero() {
		return fmt.Errorf("replace staff roles: staffID is required")
	}
	roleIDs = compactObjectIDs(roleIDs)
	now := time.Now().Unix()

	disableFilters := []common.Filter{
		common.Eq(staffRoleFieldStaffID, staffID),
		common.Active(),
	}
	if len(roleIDs) > 0 {
		disableFilters = append(disableFilters, common.Nin(staffRoleFieldRoleID, roleIDs))
	}
	if _, err := r.UpdateManyBy(ctx, common.And(disableFilters...), common.NewUpdateDoc().
		Set(staffRoleFieldStatus, commonmodel.StatusDeleted).
		Set(staffRoleFieldUpdatedAt, now).
		Inc(staffRoleFieldVersion, 1),
	); err != nil {
		return fmt.Errorf("replace staff roles: %w", err)
	}

	for _, roleID := range roleIDs {
		_, err := r.UpsertFields(ctx, bson.M{
			"staff_id": staffID,
			"role_id":  roleID,
		}, bson.M{
			"staff_id":    staffID,
			"role_id":     roleID,
			"assigned_by": assignedBy,
			"assigned_at": now,
			"status":      commonmodel.StatusActive,
		})
		if err != nil {
			return fmt.Errorf("replace staff roles: %w", err)
		}
	}
	return nil
}

func (r *StaffRoleRepository) RollbackCreateByStaffID(ctx context.Context, staffID bson.ObjectID) error {
	if staffID.IsZero() {
		return fmt.Errorf("rollback staff roles create: staffID is required")
	}
	if err := r.DeleteAllBy(ctx, common.Eq(staffRoleFieldStaffID, staffID)); err != nil {
		return fmt.Errorf("rollback staff roles create: %w", err)
	}
	return nil
}
