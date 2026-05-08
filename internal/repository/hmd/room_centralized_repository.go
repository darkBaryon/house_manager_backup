package hmd

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RoomCentralizedRepository struct {
	*common.Repository[model.HmdRoomCentralized]
}

func NewRoomCentralizedRepository(client *dbmongo.Client) *RoomCentralizedRepository {
	return &RoomCentralizedRepository{
		Repository: common.NewRepository[model.HmdRoomCentralized](client.Collection(model.CollectionHmdRoomCentralized)),
	}
}

func (r *RoomCentralizedRepository) Create(ctx context.Context, entity *model.HmdRoomCentralized) error {
	if entity == nil {
		return fmt.Errorf("create hmd room centralized: entity is nil")
	}
	if entity.RoomNo == "" || entity.RentMode == "" {
		return fmt.Errorf("create hmd room centralized: roomNo and rentMode are required")
	}
	return r.Insert(ctx, entity)
}

func (r *RoomCentralizedRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hmd room centralized by id: id is required")
	}
	return r.FindById(ctx, id)
}

func (r *RoomCentralizedRepository) FindByBuildingAndRoomNo(ctx context.Context, buildingID bson.ObjectID, roomNo string) (*model.HmdRoomCentralized, error) {
	if buildingID.IsZero() || roomNo == "" {
		return nil, fmt.Errorf("find hmd room centralized by building and roomNo: buildingID and roomNo are required")
	}
	return r.FindOne(ctx, bson.M{"building_id": buildingID, "room_no": roomNo, "status": model.StatusActive})
}

func (r *RoomCentralizedRepository) ListByBuildingID(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	if buildingID.IsZero() {
		return nil, fmt.Errorf("list hmd rooms centralized by buildingID: buildingID is required")
	}
	return r.FindMany(ctx, bson.M{"building_id": buildingID, "status": model.StatusActive})
}

func (r *RoomCentralizedRepository) ListByProjectID(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	if projectID.IsZero() {
		return nil, fmt.Errorf("list hmd rooms centralized by projectID: projectID is required")
	}
	return r.FindMany(ctx, bson.M{"project_id": projectID, "status": model.StatusActive})
}

func (r *RoomCentralizedRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd room centralized base info: id is required")
	}
	if len(fields) == 0 {
		return fmt.Errorf("update hmd room centralized base info: fields are required")
	}
	return r.UpdateFieldsById(ctx, id, fields)
}

func (r *RoomCentralizedRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, roomStatus int) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd room centralized status: id is required")
	}
	return r.UpdateFieldsById(ctx, id, bson.M{"room_status": roomStatus})
}
