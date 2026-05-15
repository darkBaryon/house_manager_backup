package hmd

import (
	"context"
	"fmt"
	hmdmodel "house-manager/internal/model/hmd"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

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

func (r *BuildingRepository) FindByBuildingCode(ctx context.Context, buildingCode string) (*hmdmodel.HmdBuilding, error) {
	if buildingCode == "" {
		return nil, fmt.Errorf("find hmd building by buildingCode: buildingCode is required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"building_code": buildingCode}))
}

func (r *BuildingRepository) ListByProjectID(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdBuilding, error) {
	if projectID.IsZero() {
		return nil, fmt.Errorf("list hmd buildings by projectID: projectID is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"project_id": projectID}), hmdListFindOptions())
}

func (r *BuildingRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd building base info: id is required")
	}
	safeFields, err := pickAllowedFields(fields, buildingBaseInfoFields)
	if err != nil {
		return fmt.Errorf("update hmd building base info: %w", err)
	}
	if err := hmdmodel.ValidateHmdUpdateFields(safeFields); err != nil {
		return fmt.Errorf("update hmd building base info: %w", err)
	}
	return r.UpdateFieldsByID(ctx, id, safeFields)
}
