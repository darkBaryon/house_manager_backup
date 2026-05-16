package landlord

import (
	"context"
	"fmt"
	"strings"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type LandlordRepository struct {
	*common.Repository[authmodel.Landlord]
}

func NewLandlordRepository(client *dbmongo.Client) *LandlordRepository {
	return &LandlordRepository{
		Repository: common.NewRepository[authmodel.Landlord](client.Collection(authmodel.CollectionLandlord)),
	}
}

func (r *LandlordRepository) Create(ctx context.Context, landlord *authmodel.Landlord) error {
	if err := landlord.ValidateForCreate(); err != nil {
		return fmt.Errorf("create landlord: %w", err)
	}
	normalizeLandlord(landlord)
	return r.Insert(ctx, landlord)
}

func (r *LandlordRepository) FindActiveByPhone(ctx context.Context, phone string) (*authmodel.Landlord, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, fmt.Errorf("find landlord by phone: phone is required")
	}
	items, err := r.FindMany(ctx, bson.M{
		"phone":  phone,
		"status": commonmodel.StatusActive,
	}, options.Find().SetLimit(2))
	if err != nil {
		return nil, fmt.Errorf("find landlord by phone: %w", err)
	}
	if len(items) > 1 {
		return nil, fmt.Errorf("find landlord by phone: multiple active landlord records found")
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func (r *LandlordRepository) FindActiveByID(ctx context.Context, id bson.ObjectID) (*authmodel.Landlord, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find landlord by id: id is required")
	}
	landlord, err := r.FindOne(ctx, bson.M{
		"_id":    id,
		"status": commonmodel.StatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("find landlord by id: %w", err)
	}
	return landlord, nil
}

func normalizeLandlord(landlord *authmodel.Landlord) {
	if landlord == nil {
		return
	}
	landlord.Phone = strings.TrimSpace(landlord.Phone)
}
