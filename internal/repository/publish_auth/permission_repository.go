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

type PermissionRepository struct {
	*common.Repository[authmodel.AdmPermission]
}

func NewPermissionRepository(client *dbmongo.Client) *PermissionRepository {
	return &PermissionRepository{
		Repository: common.NewRepository[authmodel.AdmPermission](client.Collection(authmodel.CollectionAdmPermission)),
	}
}

func (r *PermissionRepository) Create(ctx context.Context, permission *authmodel.AdmPermission) error {
	if err := permission.ValidateForCreate(); err != nil {
		return fmt.Errorf("create adm permission: %w", err)
	}
	return r.Insert(ctx, permission)
}

func (r *PermissionRepository) FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]authmodel.AdmPermission, error) {
	if len(ids) == 0 {
		return []authmodel.AdmPermission{}, nil
	}
	return r.FindMany(ctx, bson.M{
		"_id":    bson.M{"$in": ids},
		"status": commonmodel.StatusActive,
	})
}
