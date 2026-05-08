package hmd

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var buildingBaseInfoFields = map[string]struct{}{
	"building_name":      {},
	"floor_total":        {},
	"manager_name":       {},
	"manager_phone":      {},
	"photos":             {},
	"listing_facilities": {},
}

type BuildingRepository struct {
	*common.Repository[model.HmdBuilding]
}

func NewBuildingRepository(client *dbmongo.Client) *BuildingRepository {
	return &BuildingRepository{
		Repository: common.NewRepository[model.HmdBuilding](client.Collection(model.CollectionHmdBuilding)),
	}
}

func (r *BuildingRepository) Create(ctx context.Context, entity *model.HmdBuilding) error {
	if entity == nil {
		return fmt.Errorf("create hmd building: entity is nil")
	}
	if entity.ProjectID.IsZero() || entity.BuildingName == "" || entity.BuildingCode == "" {
		return fmt.Errorf("create hmd building: projectID, buildingName and buildingCode are required")
	}
	return r.Insert(ctx, entity)
}

func (r *BuildingRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hmd building by id: id is required")
	}
	return r.FindByID(ctx, id)
}

func (r *BuildingRepository) FindByBuildingCode(ctx context.Context, buildingCode string) (*model.HmdBuilding, error) {
	if buildingCode == "" {
		return nil, fmt.Errorf("find hmd building by buildingCode: buildingCode is required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"building_code": buildingCode}))
}

func (r *BuildingRepository) ListByProjectID(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error) {
	if projectID.IsZero() {
		return nil, fmt.Errorf("list hmd buildings by projectID: projectID is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"project_id": projectID}))
}

func (r *BuildingRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd building base info: id is required")
	}
	safeFields, err := pickAllowedFields(fields, buildingBaseInfoFields)
	if err != nil {
		return fmt.Errorf("update hmd building base info: %w", err)
	}
	return r.UpdateFieldsByID(ctx, id, safeFields)
}
