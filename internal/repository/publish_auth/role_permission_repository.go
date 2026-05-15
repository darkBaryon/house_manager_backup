package publishauth

import (
	"context"
	"fmt"
	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RolePermissionRepository struct {
	*common.Repository[authmodel.AdmRolePermission]
}

func NewRolePermissionRepository(client *dbmongo.Client) *RolePermissionRepository {
	return &RolePermissionRepository{
		Repository: common.NewRepository[authmodel.AdmRolePermission](client.Collection(authmodel.CollectionAdmRolePermission)),
	}
}

func (r *RolePermissionRepository) Create(ctx context.Context, rolePermission *authmodel.AdmRolePermission) error {
	if err := rolePermission.ValidateForCreate(); err != nil {
		return fmt.Errorf("create adm role permission: %w", err)
	}
	return r.Insert(ctx, rolePermission)
}

func (r *RolePermissionRepository) ListActiveByRoleIDs(ctx context.Context, roleIDs []bson.ObjectID) ([]authmodel.AdmRolePermission, error) {
	if len(roleIDs) == 0 {
		return []authmodel.AdmRolePermission{}, nil
	}
	return r.FindMany(ctx, bson.M{
		"role_id": bson.M{"$in": roleIDs},
		"status":  commonmodel.StatusActive,
	})
}
