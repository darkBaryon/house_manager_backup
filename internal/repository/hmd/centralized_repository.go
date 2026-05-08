package hmd

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var centralizedBaseInfoFields = map[string]struct{}{
	"project_name": {},
	"city":         {},
	"district":     {},
	"address_text": {},
	"geo":          {},
	"brand_name":   {},
}

type CentralizedRepository struct {
	*common.Repository[model.HmdCentralized]
}

func NewCentralizedRepository(client *dbmongo.Client) *CentralizedRepository {
	return &CentralizedRepository{
		Repository: common.NewRepository[model.HmdCentralized](client.Collection(model.CollectionHmdCentralized)),
	}
}

func (r *CentralizedRepository) Create(ctx context.Context, entity *model.HmdCentralized) error {
	if entity == nil {
		return fmt.Errorf("create hmd centralized: entity is nil")
	}
	if entity.ProjectName == "" || entity.ProjectCode == "" || entity.City == "" {
		return fmt.Errorf("create hmd centralized: projectName, projectCode and city are required")
	}
	return r.Insert(ctx, entity)
}

func (r *CentralizedRepository) FindByProjectCode(ctx context.Context, projectCode string) (*model.HmdCentralized, error) {
	if projectCode == "" {
		return nil, fmt.Errorf("find hmd centralized by projectCode: projectCode is required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"project_code": projectCode}))
}

func (r *CentralizedRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hmd centralized by id: id is required")
	}
	return r.Repository.FindByID(ctx, id)
}

func (r *CentralizedRepository) ListByCity(ctx context.Context, city string) ([]model.HmdCentralized, error) {
	if city == "" {
		return nil, fmt.Errorf("list hmd centralized by city: city is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"city": city}))
}

func (r *CentralizedRepository) ListByCityAndDistrict(ctx context.Context, city, district string) ([]model.HmdCentralized, error) {
	if city == "" || district == "" {
		return nil, fmt.Errorf("list hmd centralized by city and district: city and district are required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"city": city, "district": district}))
}

func (r *CentralizedRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd centralized base info: id is required")
	}
	safeFields, err := pickAllowedFields(fields, centralizedBaseInfoFields)
	if err != nil {
		return fmt.Errorf("update hmd centralized base info: %w", err)
	}
	return r.UpdateFieldsByID(ctx, id, safeFields)
}
