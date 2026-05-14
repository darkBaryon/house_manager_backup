package hmd

import (
	"context"
	"fmt"

	"house-manager/internal/model"
	"house-manager/internal/repository/common"
	dbmongo "house-manager/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RoomDecentralizedRepository struct {
	*common.Repository[model.HmdRoomDecentralized]
}

func NewRoomDecentralizedRepository(client *dbmongo.Client) *RoomDecentralizedRepository {
	return &RoomDecentralizedRepository{
		Repository: common.NewRepository[model.HmdRoomDecentralized](client.Collection(model.CollectionHmdRoomDecentralized)),
	}
}

func (r *RoomDecentralizedRepository) Create(ctx context.Context, entity *model.HmdRoomDecentralized) error {
	if entity != nil && entity.RoomStatus == model.RoomStatusUnspecified {
		entity.RoomStatus = model.RoomStatusAvailable
	}
	if err := entity.ValidateForCreate(); err != nil {
		return fmt.Errorf("create hmd room decentralized: %w", err)
	}
	return r.Insert(ctx, entity)
}

func (r *RoomDecentralizedRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.HmdRoomDecentralized, error) {
	if id.IsZero() {
		return nil, fmt.Errorf("find hmd room decentralized by id: id is required")
	}
	return r.Repository.FindByID(ctx, id)
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
	return r.FindMany(ctx, activeFilter(bson.M{"decentralized_id": decentralizedID}), hmdListFindOptions())
}

func (r *RoomDecentralizedRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd room decentralized base info: id is required")
	}
	safeFields, err := pickAllowedFields(fields, roomDecentralizedBaseInfoFields)
	if err != nil {
		return fmt.Errorf("update hmd room decentralized base info: %w", err)
	}
	if err := model.ValidateHmdUpdateFields(safeFields); err != nil {
		return fmt.Errorf("update hmd room decentralized base info: %w", err)
	}
	return r.UpdateFieldsByID(ctx, id, safeFields)
}

func (r *RoomDecentralizedRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, roomStatus int) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd room decentralized status: id is required")
	}
	if !model.IsValidRoomStatusUpdateTarget(roomStatus) {
		return fmt.Errorf("update hmd room decentralized status: roomStatus is invalid")
	}
	return r.UpdateFieldsByID(ctx, id, roomStatusUpdateFields(roomStatus))
}
