package hmd

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var roomTypeCentralizedBaseInfoFields = map[string]struct{}{
	"room_type_name":   {},
	"room_count":       {},
	"hall_count":       {},
	"bathroom_count":   {},
	"kitchen_count":    {},
	"area_size":        {},
	"orientation":      {},
	"decoration_level": {},
	"payment_cycle":    {},
	"rent":             {},
	"deposit":          {},
	"service_fee":      {},
	"agency_fee_mode":  {},
	"agency_fee_value": {},
	"images":           {},
	"room_facilities":  {},
}

type RoomTypeCentralizedRepository struct {
	*common.Repository[model.HmdRoomTypeCentralized]
}

func NewRoomTypeCentralizedRepository(client *dbmongo.Client) *RoomTypeCentralizedRepository {
	return &RoomTypeCentralizedRepository{
		Repository: common.NewRepository[model.HmdRoomTypeCentralized](client.Collection(model.CollectionHmdRoomTypeCentralized)),
	}
}

func (r *RoomTypeCentralizedRepository) Create(ctx context.Context, entity *model.HmdRoomTypeCentralized) error {
	if entity == nil {
		return fmt.Errorf("create hmd room type centralized: entity is nil")
	}
	if entity.RoomTypeName == "" {
		return fmt.Errorf("create hmd room type centralized: roomTypeName is required")
	}
	if entity.ProjectID.IsZero() && entity.BuildingID.IsZero() {
		return fmt.Errorf("create hmd room type centralized: projectID or buildingID is required")
	}
	return r.Insert(ctx, entity)
}

func (r *RoomTypeCentralizedRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdRoomTypeCentralized, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hmd room type centralized by id: id is required")
	}
	return r.FindByID(ctx, id)
}

func (r *RoomTypeCentralizedRepository) FindByProjectAndName(ctx context.Context, projectID bson.ObjectID, roomTypeName string) (*model.HmdRoomTypeCentralized, error) {
	if projectID.IsZero() || roomTypeName == "" {
		return nil, fmt.Errorf("find hmd room type centralized by project and name: projectID and roomTypeName are required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"project_id": projectID, "room_type_name": roomTypeName}))
}

func (r *RoomTypeCentralizedRepository) ListByProjectID(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error) {
	if projectID.IsZero() {
		return nil, fmt.Errorf("list hmd room types centralized by projectID: projectID is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"project_id": projectID}))
}

func (r *RoomTypeCentralizedRepository) ListByBuildingID(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error) {
	if buildingID.IsZero() {
		return nil, fmt.Errorf("list hmd room types centralized by buildingID: buildingID is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"building_id": buildingID}))
}

func (r *RoomTypeCentralizedRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd room type centralized base info: id is required")
	}
	safeFields, err := pickAllowedFields(fields, roomTypeCentralizedBaseInfoFields)
	if err != nil {
		return fmt.Errorf("update hmd room type centralized base info: %w", err)
	}
	return r.UpdateFieldsByID(ctx, id, safeFields)
}
