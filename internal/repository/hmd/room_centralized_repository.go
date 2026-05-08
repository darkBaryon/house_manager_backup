package hmd

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var roomCentralizedBaseInfoFields = map[string]struct{}{
	"room_no":            {},
	"floor_no":           {},
	"rent_mode":          {},
	"layout_text":        {},
	"area_size":          {},
	"orientation":        {},
	"decoration_level":   {},
	"payment_cycle":      {},
	"rent":               {},
	"deposit":            {},
	"service_fee":        {},
	"agency_fee_mode":    {},
	"agency_fee_value":   {},
	"viewing_time_rule":  {},
	"start_rent_rule":    {},
	"images":             {},
	"room_facilities":    {},
	"listing_facilities": {},
}

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
	if entity.ProjectID.IsZero() || entity.BuildingID.IsZero() || entity.RoomNo == "" || entity.RentMode == "" {
		return fmt.Errorf("create hmd room centralized: projectID, buildingID, roomNo and rentMode are required")
	}
	return r.Insert(ctx, entity)
}

func (r *RoomCentralizedRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hmd room centralized by id: id is required")
	}
	return r.FindByID(ctx, id)
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
	return r.FindMany(ctx, activeFilter(bson.M{"building_id": buildingID}))
}

func (r *RoomCentralizedRepository) ListByProjectID(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	if projectID.IsZero() {
		return nil, fmt.Errorf("list hmd rooms centralized by projectID: projectID is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"project_id": projectID}))
}

func (r *RoomCentralizedRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd room centralized base info: id is required")
	}
	safeFields, err := pickAllowedFields(fields, roomCentralizedBaseInfoFields)
	if err != nil {
		return fmt.Errorf("update hmd room centralized base info: %w", err)
	}
	return r.UpdateFieldsByID(ctx, id, safeFields)
}

func (r *RoomCentralizedRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, roomStatus int) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd room centralized status: id is required")
	}
	return r.UpdateFieldsByID(ctx, id, roomStatusUpdateFields(roomStatus))
}
