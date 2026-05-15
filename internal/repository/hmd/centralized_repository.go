package hmd

import (
	"context"
	"fmt"
	hmdmodel "house-manager/internal/model/hmd"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CentralizedRepository struct {
	*common.Repository[hmdmodel.HmdCentralized]
}

func NewCentralizedRepository(client *dbmongo.Client) *CentralizedRepository {
	return &CentralizedRepository{
		Repository: common.NewRepository[hmdmodel.HmdCentralized](client.Collection(hmdmodel.CollectionHmdCentralized)),
	}
}

func (r *CentralizedRepository) Create(ctx context.Context, entity *hmdmodel.HmdCentralized) error {
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create hmd centralized: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *CentralizedRepository) FindByProjectCode(ctx context.Context, projectCode string) (*hmdmodel.HmdCentralized, error) {
	if projectCode == "" {
		return nil, fmt.Errorf("find hmd centralized by projectCode: projectCode is required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"project_code": projectCode}))
}

func (r *CentralizedRepository) FindByID(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdCentralized, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hmd centralized by id: id is required")
	}
	return r.Repository.FindByID(ctx, id)
}

func (r *CentralizedRepository) List(ctx context.Context, city, district string) ([]hmdmodel.HmdCentralized, error) {
	filter := bson.M{}
	if city != "" {
		filter["city"] = city
	}
	if district != "" {
		filter["district"] = district
	}
	return r.FindMany(ctx, activeFilter(filter), hmdListFindOptions())
}

func (r *CentralizedRepository) ListByCity(ctx context.Context, city string) ([]hmdmodel.HmdCentralized, error) {
	if city == "" {
		return nil, fmt.Errorf("list hmd centralized by city: city is required")
	}
	return r.List(ctx, city, "")
}

func (r *CentralizedRepository) ListByCityAndDistrict(ctx context.Context, city, district string) ([]hmdmodel.HmdCentralized, error) {
	if city == "" || district == "" {
		return nil, fmt.Errorf("list hmd centralized by city and district: city and district are required")
	}
	return r.List(ctx, city, district)
}

func (r *CentralizedRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd centralized base info: id is required")
	}
	safeFields, err := pickAllowedFields(fields, centralizedBaseInfoFields)
	if err != nil {
		return fmt.Errorf("update hmd centralized base info: %w", err)
	}
	if err := hmdmodel.ValidateHmdUpdateFields(safeFields); err != nil {
		return fmt.Errorf("update hmd centralized base info: %w", err)
	}
	return r.UpdateFieldsByID(ctx, id, safeFields)
}
