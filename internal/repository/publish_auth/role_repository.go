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

type RoleRepository struct {
	*common.Repository[authmodel.AdmRole]
}

func NewRoleRepository(client *dbmongo.Client) *RoleRepository {
	return &RoleRepository{
		Repository: common.NewRepository[authmodel.AdmRole](client.Collection(authmodel.CollectionAdmRole)),
	}
}

func (r *RoleRepository) Create(ctx context.Context, role *authmodel.AdmRole) error {
	if err := role.ValidateForCreate(); err != nil {
		return fmt.Errorf("create adm role: %w", err)
	}
	return r.Insert(ctx, role)
}

func (r *RoleRepository) FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]authmodel.AdmRole, error) {
	if len(ids) == 0 {
		return []authmodel.AdmRole{}, nil
	}
	return r.FindMany(ctx, bson.M{
		"_id":    bson.M{"$in": ids},
		"status": commonmodel.StatusActive,
	})
}
