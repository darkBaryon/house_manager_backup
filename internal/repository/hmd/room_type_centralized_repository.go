package hmd

import (
	"context"
	"fmt"
	hmdmodel "house-manager/internal/model/hmd"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RoomTypeCentralizedRepository struct {
	*common.Repository[hmdmodel.HmdRoomTypeCentralized]
}

func NewRoomTypeCentralizedRepository(client *dbmongo.Client) *RoomTypeCentralizedRepository {
	return &RoomTypeCentralizedRepository{
		Repository: common.NewRepository[hmdmodel.HmdRoomTypeCentralized](client.Collection(hmdmodel.CollectionHmdRoomTypeCentralized)),
	}
}

func (r *RoomTypeCentralizedRepository) Create(ctx context.Context, entity *hmdmodel.HmdRoomTypeCentralized) error {
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create hmd room type centralized: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *RoomTypeCentralizedRepository) FindByID(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomTypeCentralized, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hmd room type centralized by id: id is required")
	}
	return r.Repository.FindByID(ctx, id)
}

func (r *RoomTypeCentralizedRepository) FindByProjectAndName(ctx context.Context, projectID bson.ObjectID, roomTypeName string) (*hmdmodel.HmdRoomTypeCentralized, error) {
	if projectID.IsZero() || roomTypeName == "" {
		return nil, fmt.Errorf("find hmd room type centralized by project and name: projectID and roomTypeName are required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"project_id": projectID, "room_type_name": roomTypeName}))
}

func (r *RoomTypeCentralizedRepository) ListByProjectID(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomTypeCentralized, error) {
	if projectID.IsZero() {
		return nil, fmt.Errorf("list hmd room types centralized by projectID: projectID is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"project_id": projectID}), hmdListFindOptions())
}

func (r *RoomTypeCentralizedRepository) ListByBuildingID(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomTypeCentralized, error) {
	if buildingID.IsZero() {
		return nil, fmt.Errorf("list hmd room types centralized by buildingID: buildingID is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"building_id": buildingID}), hmdListFindOptions())
}

func (r *RoomTypeCentralizedRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd room type centralized base info: id is required")
	}
	safeFields, err := pickAllowedFields(fields, roomTypeCentralizedBaseInfoFields)
	if err != nil {
		return fmt.Errorf("update hmd room type centralized base info: %w", err)
	}
	if err := hmdmodel.ValidateHmdUpdateFields(safeFields); err != nil {
		return fmt.Errorf("update hmd room type centralized base info: %w", err)
	}
	return r.UpdateFieldsByID(ctx, id, safeFields)
}
