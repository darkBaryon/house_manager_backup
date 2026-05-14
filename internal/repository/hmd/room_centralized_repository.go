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
	if entity != nil && entity.RoomStatus == model.RoomStatusUnspecified {
		entity.RoomStatus = model.RoomStatusAvailable
	}
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create hmd room centralized: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *RoomCentralizedRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hmd room centralized by id: id is required")
	}
	return r.Repository.FindByID(ctx, id)
}

func (r *RoomCentralizedRepository) FindByBuildingAndRoomNo(ctx context.Context, buildingID bson.ObjectID, roomNo string) (*model.HmdRoomCentralized, error) {
	if buildingID.IsZero() || roomNo == "" {
		return nil, fmt.Errorf("find hmd room centralized by building and roomNo: buildingID and roomNo are required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"building_id": buildingID, "room_no": roomNo}))
}

func (r *RoomCentralizedRepository) ListByBuildingID(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	if buildingID.IsZero() {
		return nil, fmt.Errorf("list hmd rooms centralized by buildingID: buildingID is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"building_id": buildingID}), hmdListFindOptions())
}

func (r *RoomCentralizedRepository) ListByProjectID(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	if projectID.IsZero() {
		return nil, fmt.Errorf("list hmd rooms centralized by projectID: projectID is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"project_id": projectID}), hmdListFindOptions())
}

func (r *RoomCentralizedRepository) ListByRoomTypeID(ctx context.Context, roomTypeID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	if roomTypeID.IsZero() {
		return nil, fmt.Errorf("list hmd rooms centralized by roomTypeID: roomTypeID is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"room_type_id": roomTypeID}), hmdListFindOptions())
}

func (r *RoomCentralizedRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd room centralized base info: id is required")
	}
	safeFields, err := pickAllowedFields(fields, roomCentralizedBaseInfoFields)
	if err != nil {
		return fmt.Errorf("update hmd room centralized base info: %w", err)
	}
	if err := model.ValidateHmdUpdateFields(safeFields); err != nil {
		return fmt.Errorf("update hmd room centralized base info: %w", err)
	}
	return r.UpdateFieldsByID(ctx, id, safeFields)
}

func (r *RoomCentralizedRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, roomStatus int) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd room centralized status: id is required")
	}
	if !model.IsValidRoomStatusUpdateTarget(roomStatus) {
		return fmt.Errorf("update hmd room centralized status: roomStatus is invalid")
	}
	return r.UpdateFieldsByID(ctx, id, roomStatusUpdateFields(roomStatus))
}
