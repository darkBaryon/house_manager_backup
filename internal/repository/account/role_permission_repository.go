package account

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RolePermissionRepository struct {
	*common.Repository[model.AdmRolePermission]
}

func NewRolePermissionRepository(client *dbmongo.Client) *RolePermissionRepository {
	return &RolePermissionRepository{
		Repository: common.NewRepository[model.AdmRolePermission](client.Collection(model.CollectionAdmRolePermission)),
	}
}

func (r *RolePermissionRepository) Create(ctx context.Context, rolePermission *model.AdmRolePermission) error {
	if err := rolePermission.ValidateForCreate(); err != nil {
		return fmt.Errorf("create adm role permission: %w", err)
	}
	return r.Insert(ctx, rolePermission)
}

func (r *RolePermissionRepository) ListActiveByRoleIDs(ctx context.Context, roleIDs []bson.ObjectID) ([]model.AdmRolePermission, error) {
	if len(roleIDs) == 0 {
		return []model.AdmRolePermission{}, nil
	}
	return r.FindMany(ctx, bson.M{
		"role_id": bson.M{"$in": roleIDs},
		"status":  model.StatusActive,
	})
}
