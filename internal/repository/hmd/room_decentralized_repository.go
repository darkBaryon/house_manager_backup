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
	return r.FindById(ctx, id)
}

func (r *RoomDecentralizedRepository) FindByDecentralizedAndRoomNo(ctx context.Context, decentralizedID bson.ObjectID, roomNo string) (*model.HmdRoomDecentralized, error) {
	if decentralizedID.IsZero() || roomNo == "" {
		return nil, fmt.Errorf("find hmd room decentralized by decentralizedID and roomNo: decentralizedID and roomNo are required")
	}
	return r.FindOne(ctx, bson.M{"decentralized_id": decentralizedID, "room_no": roomNo, "status": model.StatusActive})
}

func (r *RoomDecentralizedRepository) ListByDecentralizedID(ctx context.Context, decentralizedID bson.ObjectID) ([]model.HmdRoomDecentralized, error) {
	if decentralizedID.IsZero() {
		return nil, fmt.Errorf("list hmd rooms decentralized by decentralizedID: decentralizedID is required")
	}
	return r.FindMany(ctx, bson.M{"decentralized_id": decentralizedID, "status": model.StatusActive})
}

func (r *RoomDecentralizedRepository) UpdateBaseInfo(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd room decentralized base info: id is required")
	}
	if len(fields) == 0 {
		return fmt.Errorf("update hmd room decentralized base info: fields are required")
	}
	return r.UpdateFieldsById(ctx, id, fields)
}

func (r *RoomDecentralizedRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, roomStatus int) error {
	if id.IsZero() {
		return fmt.Errorf("update hmd room decentralized status: id is required")
	}
	return r.UpdateFieldsById(ctx, id, bson.M{"room_status": roomStatus})
}
