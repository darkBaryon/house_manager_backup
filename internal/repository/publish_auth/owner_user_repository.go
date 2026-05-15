package publishauth

import (
	"context"
	"fmt"
	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type OwnerUserRepository struct {
	*common.Repository[authmodel.User]
}

func NewOwnerUserRepository(client *dbmongo.Client) *OwnerUserRepository {
	return &OwnerUserRepository{
		Repository: common.NewRepository[authmodel.User](client.Collection(authmodel.CollectionUser)),
	}
}

func (r *OwnerUserRepository) FindActiveByPhone(ctx context.Context, phone string) (*authmodel.User, error) {
	if phone == "" {
		return nil, fmt.Errorf("find owner user by phone: phone is required")
	}
	items, err := r.FindMany(ctx, bson.M{
		"phone":  phone,
		"status": commonmodel.StatusActive,
	}, options.Find().SetLimit(2))
	if err != nil {
		return nil, fmt.Errorf("find owner user by phone: %w", err)
	}
	if len(items) > 1 {
		return nil, fmt.Errorf("find owner user by phone: multiple active user records found")
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func (r *OwnerUserRepository) FindByID(ctx context.Context, id bson.ObjectID) (*authmodel.User, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find owner user by id: id is required")
	}
	user, err := r.FindOne(ctx, bson.M{
		"_id":    id,
		"status": commonmodel.StatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("find owner user by id: %w", err)
	}
	return user, nil
}
