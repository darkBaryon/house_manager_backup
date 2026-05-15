package publishauth

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *StaffRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "phone", Value: 1}},
			Options: options.Index().SetName("phone_1"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("status_1_updated_at_-1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure adm staff indexes: %w", err)
	}
	return nil
}

func (r *RoleRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "role_code", Value: 1}},
			Options: options.Index().SetName("role_code_1").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("status_1_updated_at_-1"),
		},
	}
	if _, err := r.Collection.Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("ensure adm role indexes: %w", err)
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
		return fmt.Errorf("ensure adm permission indexes: %w", err)
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
		return fmt.Errorf("ensure adm staff role indexes: %w", err)
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
		return fmt.Errorf("ensure adm role permission indexes: %w", err)
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
		return fmt.Errorf("ensure adm login log indexes: %w", err)
	}
	return nil
}
