package adm

import (
	"context"
	"fmt"

	"house-manager/internal/repository/common"
)

func (r *StaffRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex("phone_1", common.IndexKey(staffFieldPhone, common.SortAsc)).WithUnique(),
		common.NewIndex(
			"status_1_updated_at_-1",
			common.IndexKey(staffFieldStatus, common.SortAsc),
			common.IndexKey(staffFieldUpdatedAt, common.SortDesc),
		),
	); err != nil {
		return fmt.Errorf("ensure admin staff indexes: %w", err)
	}
	return nil
}

func (r *StaffAuthRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex("staff_id_1", common.IndexKey(staffAuthFieldStaffID, common.SortAsc)).WithUnique(),
		common.NewIndex(
			"auth_type_1_status_1",
			common.IndexKey(staffAuthFieldAuthType, common.SortAsc),
			common.IndexKey(staffAuthFieldStatus, common.SortAsc),
		),
	); err != nil {
		return fmt.Errorf("ensure admin staff auth indexes: %w", err)
	}
	return nil
}

func (r *RoleRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex("role_code_1", common.IndexKey(roleFieldRoleCode, common.SortAsc)).WithUnique(),
		common.NewIndex(
			"status_1_updated_at_-1",
			common.IndexKey(roleFieldStatus, common.SortAsc),
			common.IndexKey(roleFieldUpdatedAt, common.SortDesc),
		),
	); err != nil {
		return fmt.Errorf("ensure admin role indexes: %w", err)
	}
	return nil
}

func (r *PermissionRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex("permission_code_1", common.IndexKey(permissionFieldPermissionCode, common.SortAsc)).WithUnique(),
		common.NewIndex(
			"module_1_status_1",
			common.IndexKey(permissionFieldModule, common.SortAsc),
			common.IndexKey(permissionFieldStatus, common.SortAsc),
		),
	); err != nil {
		return fmt.Errorf("ensure admin permission indexes: %w", err)
	}
	return nil
}

func (r *StaffRoleRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex(
			"staff_id_1_role_id_1",
			common.IndexKey(staffRoleFieldStaffID, common.SortAsc),
			common.IndexKey(staffRoleFieldRoleID, common.SortAsc),
		).WithUnique(),
		common.NewIndex(
			"role_id_1_status_1",
			common.IndexKey(staffRoleFieldRoleID, common.SortAsc),
			common.IndexKey(staffRoleFieldStatus, common.SortAsc),
		),
	); err != nil {
		return fmt.Errorf("ensure admin staff role indexes: %w", err)
	}
	return nil
}

func (r *RolePermissionRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex(
			"role_id_1_permission_id_1",
			common.IndexKey(rolePermissionFieldRoleID, common.SortAsc),
			common.IndexKey(rolePermissionFieldPermissionID, common.SortAsc),
		).WithUnique(),
		common.NewIndex(
			"permission_id_1_status_1",
			common.IndexKey(rolePermissionFieldPermissionID, common.SortAsc),
			common.IndexKey(rolePermissionFieldStatus, common.SortAsc),
		),
	); err != nil {
		return fmt.Errorf("ensure admin role permission indexes: %w", err)
	}
	return nil
}

func (r *LoginLogRepository) EnsureIndexes(ctx context.Context) error {
	if err := r.Repository.EnsureIndexes(ctx,
		common.NewIndex(
			"staff_id_1_login_at_-1",
			common.IndexKey(loginLogFieldStaffID, common.SortAsc),
			common.IndexKey(loginLogFieldLoginAt, common.SortDesc),
		),
		common.NewIndex("login_at_-1", common.IndexKey(loginLogFieldLoginAt, common.SortDesc)),
	); err != nil {
		return fmt.Errorf("ensure admin login log indexes: %w", err)
	}
	return nil
}
