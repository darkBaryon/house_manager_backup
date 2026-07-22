package hmd

import (
	"context"
	"fmt"
	hmdmodel "house-manager/internal/model/hmd"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type BuildingRepository struct {
	*common.Repository[hmdmodel.HmdBuilding]
}

func NewBuildingRepository(client *dbmongo.Client) *BuildingRepository {
	return &BuildingRepository{
		Repository: common.NewRepository[hmdmodel.HmdBuilding](client.Collection(hmdmodel.CollectionHmdBuilding)),
	}
}

func (r *BuildingRepository) Create(ctx context.Context, entity *hmdmodel.HmdBuilding) error {
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create hmd building: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *BuildingRepository) FindByID(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdBuilding, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hmd building by id: id is required")
	}
	return r.Repository.FindByID(ctx, id)
}

func (r *BuildingRepository) FindByProjectAndName(ctx context.Context, projectID bson.ObjectID, buildingName string) (*hmdmodel.HmdBuilding, error) {
	if projectID.IsZero() {
		return nil, fmt.Errorf("find hmd building by project and name: projectID is required")
	}
	buildingName = strings.TrimSpace(buildingName)
	if buildingName == "" {
		return nil, fmt.Errorf("find hmd building by project and name: buildingName is required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"project_id": projectID, "building_name": buildingName}))
}

func (r *BuildingRepository) ListByProjectID(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdBuilding, error) {
	if projectID.IsZero() {
		return nil, fmt.Errorf("list hmd buildings by projectID: projectID is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"project_id": projectID}), hmdListFindOptions())
}

func (r *BuildingRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, update BuildingBaseInfoUpdate) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd building base info: id is required")
	}
	fields, err := buildingBaseInfoUpdateFields(update)
	if err != nil {
		return fmt.Errorf("update hmd building base info: %w", err)
	}
	return r.UpdateFieldsByID(ctx, id, fields)
}
