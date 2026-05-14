package account

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RoleRepository struct {
	*common.Repository[model.AdmRole]
}

func NewRoleRepository(client *dbmongo.Client) *RoleRepository {
	return &RoleRepository{
		Repository: common.NewRepository[model.AdmRole](client.Collection(model.CollectionAdmRole)),
	}
}

func (r *RoleRepository) Create(ctx context.Context, role *model.AdmRole) error {
	if err := role.ValidateForCreate(); err != nil {
		return fmt.Errorf("create adm role: %w", err)
	}
	return r.Insert(ctx, role)
}

func (r *RoleRepository) FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]model.AdmRole, error) {
	if len(ids) == 0 {
		return []model.AdmRole{}, nil
	}
	return r.FindMany(ctx, bson.M{
		"_id":    bson.M{"$in": ids},
		"status": model.StatusActive,
	})
}
