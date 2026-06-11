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

type RolePermissionRepository struct {
	*common.Repository[authmodel.AdmRolePermission]
	core *relationcore.Core[authmodel.AdmRolePermission]
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
	repo := common.NewRepository[authmodel.AdmRolePermission](client.Collection(authmodel.CollectionAdmRolePermission))
	return &RolePermissionRepository{
		Repository: repo,
		core: relationcore.NewCore(repo, relationcore.Config[authmodel.AdmRolePermission]{
			LeftField:       rolePermissionFieldRoleID,
			RightField:      rolePermissionFieldPermissionID,
			AssignedByField: rolePermissionFieldAssignedBy,
			AssignedAtField: rolePermissionFieldAssignedAt,
			Validate:        (*authmodel.AdmRolePermission).ValidateForCreate,
		}),
	}
}

func (r *RolePermissionRepository) ListActiveByRoleIDs(ctx context.Context, roleIDs []bson.ObjectID) ([]authmodel.AdmRolePermission, error) {
	items, err := r.core.ListActiveByLefts(ctx, roleIDs)
	if err != nil {
		return nil, fmt.Errorf("list role permissions: %w", err)
	}
	return items, nil
}

func (r *RolePermissionRepository) CreateMany(ctx context.Context, roleID bson.ObjectID, permissionIDs []bson.ObjectID, assignedBy bson.ObjectID) error {
	if roleID.IsZero() {
		return fmt.Errorf("create role permissions: roleID is required")
	}
	if err := r.core.CreateMany(ctx, roleID, permissionIDs, assignedBy); err != nil {
		return fmt.Errorf("create role permissions: %w", err)
	}
	return nil
}

func (r *RolePermissionRepository) ReplaceByRoleID(ctx context.Context, roleID bson.ObjectID, permissionIDs []bson.ObjectID, assignedBy bson.ObjectID) error {
	if roleID.IsZero() {
		return fmt.Errorf("replace role permissions: roleID is required")
	}
	if err := r.core.ReplaceByLeft(ctx, roleID, permissionIDs, assignedBy); err != nil {
		return fmt.Errorf("replace role permissions: %w", err)
	}
	return nil
}

func (r *RolePermissionRepository) RollbackCreateByRoleID(ctx context.Context, roleID bson.ObjectID) error {
	if roleID.IsZero() {
		return fmt.Errorf("rollback role permissions create: roleID is required")
	}
	if err := r.core.RollbackByLeft(ctx, roleID); err != nil {
		return fmt.Errorf("rollback role permissions create: %w", err)
	}
	return nil
}
