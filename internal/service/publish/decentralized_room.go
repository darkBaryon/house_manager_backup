package publish

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	"house-manager/pkg/applog"
	"log/slog"
)

type decentralizedRoomService struct {
	hmd       decentralizedRoomDomain
	publisher mutationPublisher
	access    publishAccessService
}

func newDecentralizedRoomService(hmd decentralizedRoomDomain, publisher mutationPublisher, access publishAccessService) *decentralizedRoomService {
	return &decentralizedRoomService{hmd: hmd, publisher: publisher, access: access}
}

func (s *decentralizedRoomService) CreateDecentralizedRoom(ctx context.Context, input CreateDecentralizedRoomInput) (*hmdmodel.HmdRoomDecentralized, error) {
	slog.InfoContext(ctx, "publish.decentralized_room.create.start", "community_id", input.DecentralizedID.Hex(), "room_no", input.RoomNo)
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		applog.Result(ctx, "publish.decentralized_room.create.success", "publish.decentralized_room.create.failed", err)
		return nil, err
	}
	if err := requireDecentralizedCommunityAccess(ctx, scope, input.DecentralizedID, "create decentralized room"); err != nil {
		slog.WarnContext(ctx, "publish.decentralized_room.create.denied", "community_id", input.DecentralizedID.Hex(), "error", err)
		return nil, err
	}
	result, err := s.hmd.CreateDecentralizedRoom(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	if err != nil {
		applog.Result(ctx, "publish.decentralized_room.create.success", "publish.decentralized_room.create.failed", err, "community_id", input.DecentralizedID.Hex(), "room_no", input.RoomNo)
		return nil, err
	}
	slog.InfoContext(ctx, "publish.decentralized_room.create.success", "room_id", entity.ID.Hex(), "community_id", entity.DecentralizedID.Hex(), "room_no", entity.RoomNo)
	return entity, nil
}

func (s *decentralizedRoomService) GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomDecentralized, error) {
	slog.InfoContext(ctx, "publish.decentralized_room.detail.start", "room_id", id.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		applog.Result(ctx, "publish.decentralized_room.detail.success", "publish.decentralized_room.detail.failed", err, "room_id", id.Hex())
		return nil, err
	}
	room, err := s.hmd.GetDecentralizedRoom(ctx, id)
	if err != nil {
		applog.Result(ctx, "publish.decentralized_room.detail.success", "publish.decentralized_room.detail.failed", err, "room_id", id.Hex())
		return nil, err
	}
	if room != nil {
		if err := requireDecentralizedCommunityAccess(ctx, scope, room.DecentralizedID, "get decentralized room"); err != nil {
			slog.WarnContext(ctx, "publish.decentralized_room.detail.denied", "room_id", id.Hex(), "community_id", room.DecentralizedID.Hex(), "error", err)
			return nil, err
		}
	}
	slog.InfoContext(ctx, "publish.decentralized_room.detail.success", "room_id", id.Hex(), "found", room != nil)
	return room, nil
}

func (s *decentralizedRoomService) ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]hmdmodel.HmdRoomDecentralized, error) {
	slog.InfoContext(ctx, "publish.decentralized_room.list_by_community.start", "community_id", decentralizedID.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		applog.Result(ctx, "publish.decentralized_room.list_by_community.success", "publish.decentralized_room.list_by_community.failed", err, "community_id", decentralizedID.Hex())
		return nil, err
	}
	allowed, err := canAccessDecentralizedCommunity(ctx, scope, decentralizedID)
	if err != nil {
		applog.Result(ctx, "publish.decentralized_room.list_by_community.success", "publish.decentralized_room.list_by_community.failed", err, "community_id", decentralizedID.Hex())
		return nil, err
	}
	if !allowed {
		slog.WarnContext(ctx, "publish.decentralized_room.list_by_community.denied", "community_id", decentralizedID.Hex())
		return []hmdmodel.HmdRoomDecentralized{}, nil
	}
	rooms, err := s.hmd.ListDecentralizedRoomsByCommunity(ctx, decentralizedID)
	applog.Result(ctx, "publish.decentralized_room.list_by_community.success", "publish.decentralized_room.list_by_community.failed", err, "community_id", decentralizedID.Hex(), "result_count", len(rooms))
	return rooms, err
}

func (s *decentralizedRoomService) UpdateDecentralizedRoom(ctx context.Context, input UpdateDecentralizedRoomInput) (*hmdmodel.HmdRoomDecentralized, error) {
	slog.InfoContext(ctx, "publish.decentralized_room.update.start", "room_id", input.ID.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		applog.Result(ctx, "publish.decentralized_room.update.success", "publish.decentralized_room.update.failed", err, "room_id", input.ID.Hex())
		return nil, err
	}
	room, err := s.hmd.GetDecentralizedRoom(ctx, input.ID)
	if err != nil {
		applog.Result(ctx, "publish.decentralized_room.update.success", "publish.decentralized_room.update.failed", err, "room_id", input.ID.Hex())
		return nil, err
	}
	if room == nil {
		err := scopeNotFound("update decentralized room")
		slog.WarnContext(ctx, "publish.decentralized_room.update.denied", "room_id", input.ID.Hex(), "error", err)
		return nil, err
	}
	if err := requireDecentralizedCommunityAccess(ctx, scope, room.DecentralizedID, "update decentralized room"); err != nil {
		slog.WarnContext(ctx, "publish.decentralized_room.update.denied", "room_id", input.ID.Hex(), "community_id", room.DecentralizedID.Hex(), "error", err)
		return nil, err
	}
	result, err := s.hmd.UpdateDecentralizedRoom(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	applog.Result(ctx, "publish.decentralized_room.update.success", "publish.decentralized_room.update.failed", err, "room_id", input.ID.Hex())
	return entity, err
}

func (s *decentralizedRoomService) UpdateDecentralizedRoomStatus(ctx context.Context, input UpdateDecentralizedRoomStatusInput) (*hmdmodel.HmdRoomDecentralized, error) {
	slog.InfoContext(ctx, "publish.decentralized_room.update_status.start", "room_id", input.ID.Hex(), "room_status", input.RoomStatus)
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		applog.Result(ctx, "publish.decentralized_room.update_status.success", "publish.decentralized_room.update_status.failed", err, "room_id", input.ID.Hex())
		return nil, err
	}
	room, err := s.hmd.GetDecentralizedRoom(ctx, input.ID)
	if err != nil {
		applog.Result(ctx, "publish.decentralized_room.update_status.success", "publish.decentralized_room.update_status.failed", err, "room_id", input.ID.Hex())
		return nil, err
	}
	if room == nil {
		err := scopeNotFound("update decentralized room status")
		slog.WarnContext(ctx, "publish.decentralized_room.update_status.denied", "room_id", input.ID.Hex(), "error", err)
		return nil, err
	}
	if err := requireDecentralizedCommunityAccess(ctx, scope, room.DecentralizedID, "update decentralized room status"); err != nil {
		slog.WarnContext(ctx, "publish.decentralized_room.update_status.denied", "room_id", input.ID.Hex(), "community_id", room.DecentralizedID.Hex(), "error", err)
		return nil, err
	}
	result, err := s.hmd.UpdateDecentralizedRoomStatus(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	applog.Result(ctx, "publish.decentralized_room.update_status.success", "publish.decentralized_room.update_status.failed", err, "room_id", input.ID.Hex(), "room_status", input.RoomStatus)
	return entity, err
}
