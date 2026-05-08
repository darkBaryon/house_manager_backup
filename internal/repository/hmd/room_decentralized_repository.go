package hmd

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var roomDecentralizedBaseInfoFields = map[string]struct{}{
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

type RoomDecentralizedRepository struct {
	*common.Repository[model.HmdRoomDecentralized]
}

func NewRoomDecentralizedRepository(client *dbmongo.Client) *RoomDecentralizedRepository {
	return &RoomDecentralizedRepository{
		Repository: common.NewRepository[model.HmdRoomDecentralized](client.Collection(model.CollectionHmdRoomDecentralized)),
	}
}

func (r *RoomDecentralizedRepository) Create(ctx context.Context, entity *model.HmdRoomDecentralized) error {
	if entity == nil {
		return fmt.Errorf("create hmd room decentralized: entity is nil")
	}
	if entity.DecentralizedID.IsZero() || entity.RoomNo == "" || entity.RentMode == "" {
		return fmt.Errorf("create hmd room decentralized: decentralizedID, roomNo and rentMode are required")
	}
	return r.Insert(ctx, entity)
}

func (r *RoomDecentralizedRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdRoomDecentralized, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hmd room decentralized by id: id is required")
	}
	return r.FindByID(ctx, id)
}

func (r *RoomDecentralizedRepository) FindByDecentralizedAndRoomNo(ctx context.Context, decentralizedID bson.ObjectID, roomNo string) (*model.HmdRoomDecentralized, error) {
	if decentralizedID.IsZero() || roomNo == "" {
		return nil, fmt.Errorf("find hmd room decentralized by decentralizedID and roomNo: decentralizedID and roomNo are required")
	}
	return r.FindOne(ctx, activeFilter(bson.M{"decentralized_id": decentralizedID, "room_no": roomNo}))
}

func (r *RoomDecentralizedRepository) ListByDecentralizedID(ctx context.Context, decentralizedID bson.ObjectID) ([]model.HmdRoomDecentralized, error) {
	if decentralizedID.IsZero() {
		return nil, fmt.Errorf("list hmd rooms decentralized by decentralizedID: decentralizedID is required")
	}
	return r.FindMany(ctx, activeFilter(bson.M{"decentralized_id": decentralizedID}))
}

func (r *RoomDecentralizedRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd room decentralized base info: id is required")
	}
	safeFields, err := pickAllowedFields(fields, roomDecentralizedBaseInfoFields)
	if err != nil {
		return fmt.Errorf("update hmd room decentralized base info: %w", err)
	}
	return r.UpdateFieldsByID(ctx, id, safeFields)
}

func (r *RoomDecentralizedRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, roomStatus int) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd room decentralized status: id is required")
	}
	return r.UpdateFieldsByID(ctx, id, roomStatusUpdateFields(roomStatus))
}
