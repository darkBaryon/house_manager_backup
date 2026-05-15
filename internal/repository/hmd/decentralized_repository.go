package hmd

import (
	"context"
	"fmt"
	hmdmodel "house-manager/internal/model/hmd"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type DecentralizedRepository struct {
	*common.Repository[hmdmodel.HmdDecentralized]
}

func NewDecentralizedRepository(client *dbmongo.Client) *DecentralizedRepository {
	return &DecentralizedRepository{
		Repository: common.NewRepository[hmdmodel.HmdDecentralized](client.Collection(hmdmodel.CollectionHmdDecentralized)),
	}
}

func (r *DecentralizedRepository) Create(ctx context.Context, entity *hmdmodel.HmdDecentralized) error {
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create hmd decentralized: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *DecentralizedRepository) FindByID(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdDecentralized, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hmd decentralized by id: id is required")
	}
	return r.Repository.FindByID(ctx, id)
}

func (r *DecentralizedRepository) FindByCommunity(ctx context.Context, city, district, communityName string) (*hmdmodel.HmdDecentralized, error) {
	if city == "" || communityName == "" {
		return nil, fmt.Errorf("find hmd decentralized by community: city and communityName are required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{
		"city":           city,
		"district":       district,
		"community_name": communityName,
	}))
}

func (r *DecentralizedRepository) List(ctx context.Context, city, district string) ([]hmdmodel.HmdDecentralized, error) {
	filter := bson.M{}
	if city != "" {
		filter["city"] = city
	}
	if district != "" {
		filter["district"] = district
	}
	return r.FindMany(ctx, activeFilter(filter), hmdListFindOptions())
}

func (r *DecentralizedRepository) ListByCity(ctx context.Context, city string) ([]hmdmodel.HmdDecentralized, error) {
	if city == "" {
		return nil, fmt.Errorf("list hmd decentralized by city: city is required")
	}
	return r.List(ctx, city, "")
}

func (r *DecentralizedRepository) ListByDistrict(ctx context.Context, city, district string) ([]hmdmodel.HmdDecentralized, error) {
	if city == "" || district == "" {
		return nil, fmt.Errorf("list hmd decentralized by district: city and district are required")
	}
	return r.List(ctx, city, district)
}

func (r *DecentralizedRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd decentralized base info: id is required")
	}
	safeFields, err := pickAllowedFields(fields, decentralizedBaseInfoFields)
	if err != nil {
		return fmt.Errorf("update hmd decentralized base info: %w", err)
	}
	if err := hmdmodel.ValidateHmdUpdateFields(safeFields); err != nil {
		return fmt.Errorf("update hmd decentralized base info: %w", err)
	}
	return r.UpdateFieldsByID(ctx, id, safeFields)
}
