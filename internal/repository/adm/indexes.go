package adm

import (
	"context"
	"fmt"

	"house-manager/internal/repository/common"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *StaffRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "phone", Value: 1}},
			Options: options.Index().SetName("phone_1").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("status_1_updated_at_-1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure admin staff indexes: %w", err)
	}
	return nil
}

func (r *StaffAuthRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "staff_id", Value: 1}},
			Options: options.Index().SetName("staff_id_1").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "auth_type", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("auth_type_1_status_1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
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
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "permission_code", Value: 1}},
			Options: options.Index().SetName("permission_code_1").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "module", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("module_1_status_1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure admin permission indexes: %w", err)
	}
	return nil
}

func (r *StaffRoleRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "staff_id", Value: 1}, {Key: "role_id", Value: 1}},
			Options: options.Index().SetName("staff_id_1_role_id_1").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "role_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("role_id_1_status_1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure admin staff role indexes: %w", err)
	}
	return nil
}

func (r *RolePermissionRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "role_id", Value: 1}, {Key: "permission_id", Value: 1}},
			Options: options.Index().SetName("role_id_1_permission_id_1").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "permission_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("permission_id_1_status_1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure admin role permission indexes: %w", err)
	}
	return nil
}

func (r *LoginLogRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "staff_id", Value: 1}, {Key: "login_at", Value: -1}},
			Options: options.Index().SetName("staff_id_1_login_at_-1"),
		},
		{
			Keys:    bson.D{{Key: "login_at", Value: -1}},
			Options: options.Index().SetName("login_at_-1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure admin login log indexes: %w", err)
	}
	return nil
}
