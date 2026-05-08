package hmd

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

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
	if entity.ProjectID.IsZero() || entity.BuildingName == "" {
		return fmt.Errorf("create hmd building: projectID and buildingName are required")
	}
	return r.Insert(ctx, entity)
}

func (r *BuildingRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hmd building by id: id is required")
	}
	return r.FindById(ctx, id)
}

func (r *BuildingRepository) FindByBuildingCode(ctx context.Context, buildingCode string) (*model.HmdBuilding, error) {
	if buildingCode == "" {
		return nil, fmt.Errorf("find hmd building by buildingCode: buildingCode is required")
	}
	return r.FindOne(ctx, bson.M{"building_code": buildingCode, "status": model.StatusActive})
}

func (r *BuildingRepository) ListByProjectID(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error) {
	if projectID.IsZero() {
		return nil, fmt.Errorf("list hmd buildings by projectID: projectID is required")
	}
	return r.FindMany(ctx, bson.M{"project_id": projectID, "status": model.StatusActive})
}

func (r *BuildingRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd building base info: id is required")
	}
	if len(fields) == 0 {
		return fmt.Errorf("update hmd building base info: fields are required")
	}
	return r.UpdateFieldsById(ctx, id, fields)
}
