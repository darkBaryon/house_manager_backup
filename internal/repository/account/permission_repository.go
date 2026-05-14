package account

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PermissionRepository struct {
	*common.Repository[model.AdmPermission]
}

func NewPermissionRepository(client *dbmongo.Client) *PermissionRepository {
	return &PermissionRepository{
		Repository: common.NewRepository[model.AdmPermission](client.Collection(model.CollectionAdmPermission)),
	}
}

func (r *PermissionRepository) Create(ctx context.Context, permission *model.AdmPermission) error {
	if err := permission.ValidateForCreate(); err != nil {
		return fmt.Errorf("create adm permission: %w", err)
	}
	return r.Insert(ctx, permission)
}

func (r *PermissionRepository) FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]model.AdmPermission, error) {
	if len(ids) == 0 {
		return []model.AdmPermission{}, nil
	}
	return r.FindMany(ctx, bson.M{
		"_id":    bson.M{"$in": ids},
		"status": model.StatusActive,
	})
}
