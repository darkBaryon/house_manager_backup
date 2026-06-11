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

type RolePermissionRepository struct {
	*common.Repository[authmodel.AdmRolePermission]
}

const (
	rolePermissionFieldRoleID       common.Field = "role_id"
	rolePermissionFieldPermissionID common.Field = "permission_id"
	rolePermissionFieldStatus       common.Field = "status"
	rolePermissionFieldUpdatedAt    common.Field = "updated_at"
	rolePermissionFieldVersion      common.Field = "version"
	rolePermissionFieldAssignedBy   common.Field = "assigned_by"
	rolePermissionFieldAssignedAt   common.Field = "assigned_at"
)

func NewRolePermissionRepository(client *dbmongo.Client) *RolePermissionRepository {
	return &RolePermissionRepository{
		Repository: common.NewRepository[authmodel.AdmRolePermission](client.Collection(authmodel.CollectionAdmRolePermission)),
	}
}

func (r *RolePermissionRepository) ListActiveByRoleIDs(ctx context.Context, roleIDs []bson.ObjectID) ([]authmodel.AdmRolePermission, error) {
	objectIDs := compactObjectIDs(roleIDs)
	if len(objectIDs) == 0 {
		return []authmodel.AdmRolePermission{}, nil
	}
	items, err := r.FindMany(ctx, bson.M{
		"role_id": bson.M{"$in": objectIDs},
		"status":  commonmodel.StatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("list role permissions: %w", err)
	}
	return items, nil
}

func (r *RolePermissionRepository) CreateMany(ctx context.Context, roleID bson.ObjectID, permissionIDs []bson.ObjectID, assignedBy bson.ObjectID) error {
	if roleID.IsZero() {
		return fmt.Errorf("create role permissions: roleID is required")
	}
	permissionIDs = compactObjectIDs(permissionIDs)
	if len(permissionIDs) == 0 {
		return nil
	}
	now := time.Now().Unix()
	for _, permissionID := range permissionIDs {
		item := &authmodel.AdmRolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
			AssignedBy:   assignedBy,
			AssignedAt:   now,
		}
		if err := item.ValidateForCreate(); err != nil {
			return fmt.Errorf("create role permissions: %w", err)
		}
		if err := r.Insert(ctx, item); err != nil {
			return fmt.Errorf("create role permissions: %w", err)
		}
	}
	return nil
}

func (r *RolePermissionRepository) ReplaceByRoleID(ctx context.Context, roleID bson.ObjectID, permissionIDs []bson.ObjectID, assignedBy bson.ObjectID) error {
	if roleID.IsZero() {
		return fmt.Errorf("replace role permissions: roleID is required")
	}
	permissionIDs = compactObjectIDs(permissionIDs)
	now := time.Now().Unix()

	disableFilters := []common.Filter{
		common.Eq(rolePermissionFieldRoleID, roleID),
		common.Active(),
	}
	if len(permissionIDs) > 0 {
		disableFilters = append(disableFilters, common.Nin(rolePermissionFieldPermissionID, permissionIDs))
	}
	if _, err := r.UpdateManyBy(ctx, common.And(disableFilters...), common.NewUpdateDoc().
		Set(rolePermissionFieldStatus, commonmodel.StatusDeleted).
		Set(rolePermissionFieldUpdatedAt, now).
		Inc(rolePermissionFieldVersion, 1),
	); err != nil {
		return fmt.Errorf("replace role permissions: %w", err)
	}

	for _, permissionID := range permissionIDs {
		_, err := r.UpsertFields(ctx, bson.M{
			"role_id":       roleID,
			"permission_id": permissionID,
		}, bson.M{
			"role_id":       roleID,
			"permission_id": permissionID,
			"assigned_by":   assignedBy,
			"assigned_at":   now,
			"status":        commonmodel.StatusActive,
		})
		if err != nil {
			return fmt.Errorf("replace role permissions: %w", err)
		}
	}
	return nil
}

func (r *RolePermissionRepository) RollbackCreateByRoleID(ctx context.Context, roleID bson.ObjectID) error {
	if roleID.IsZero() {
		return fmt.Errorf("rollback role permissions create: roleID is required")
	}
	if err := r.DeleteAllBy(ctx, common.Eq(rolePermissionFieldRoleID, roleID)); err != nil {
		return fmt.Errorf("rollback role permissions create: %w", err)
	}
	return nil
}
