package hmd

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var decentralizedBaseInfoFields = map[string]struct{}{
	"community_name": {},
	"city":           {},
	"district":       {},
	"biz_area":       {},
	"address_text":   {},
	"geo":            {},
	"subway_station": {},
}

type DecentralizedRepository struct {
	*common.Repository[model.HmdDecentralized]
}

func NewDecentralizedRepository(client *dbmongo.Client) *DecentralizedRepository {
	return &DecentralizedRepository{
		Repository: common.NewRepository[model.HmdDecentralized](client.Collection(model.CollectionHmdDecentralized)),
	}
}

func (r *DecentralizedRepository) Create(ctx context.Context, entity *model.HmdDecentralized) error {
	if entity == nil {
		return fmt.Errorf("create hmd decentralized: entity is nil")
	}
	if entity.CommunityName == "" || entity.City == "" || entity.District == "" {
		return fmt.Errorf("create hmd decentralized: communityName, city and district are required")
	}
	return r.Insert(ctx, entity)
}

func (r *DecentralizedRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdDecentralized, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hmd decentralized by id: id is required")
	}
	return r.Repository.FindByID(ctx, id)
}

func (r *DecentralizedRepository) FindByCommunity(ctx context.Context, city, district, communityName string) (*model.HmdDecentralized, error) {
	if city == "" || district == "" || communityName == "" {
		return nil, fmt.Errorf("find hmd decentralized by community: city, district and communityName are required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{
		"city":           city,
		"district":       district,
		"community_name": communityName,
	}))
}

func (r *DecentralizedRepository) ListByCity(ctx context.Context, city string) ([]model.HmdDecentralized, error) {
	if city == "" {
		return nil, fmt.Errorf("list hmd decentralized by city: city is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"city": city}))
}

func (r *DecentralizedRepository) ListByDistrict(ctx context.Context, city, district string) ([]model.HmdDecentralized, error) {
	if city == "" || district == "" {
		return nil, fmt.Errorf("list hmd decentralized by district: city and district are required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"city": city, "district": district}))
}

func (r *DecentralizedRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd decentralized base info: id is required")
	}
	safeFields, err := pickAllowedFields(fields, decentralizedBaseInfoFields)
	if err != nil {
		return fmt.Errorf("update hmd decentralized base info: %w", err)
	}
	return r.UpdateFieldsByID(ctx, id, safeFields)
}
