package publish

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
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
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	if err := requireDecentralizedCommunityAccess(ctx, scope, input.DecentralizedID, "create decentralized room"); err != nil {
		return nil, err
	}
	result, err := s.hmd.CreateDecentralizedRoom(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *decentralizedRoomService) GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomDecentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	room, err := s.hmd.GetDecentralizedRoom(ctx, id)
	if err != nil {
		return nil, err
	}
	if room != nil {
		if err := requireDecentralizedCommunityAccess(ctx, scope, room.DecentralizedID, "get decentralized room"); err != nil {
			return nil, err
		}
	}
	return room, nil
}

func (s *decentralizedRoomService) ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]hmdmodel.HmdRoomDecentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	allowed, err := canAccessDecentralizedCommunity(ctx, scope, decentralizedID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return []hmdmodel.HmdRoomDecentralized{}, nil
	}
	return s.hmd.ListDecentralizedRoomsByCommunity(ctx, decentralizedID)
}

func (s *decentralizedRoomService) UpdateDecentralizedRoom(ctx context.Context, input UpdateDecentralizedRoomInput) (*hmdmodel.HmdRoomDecentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	room, err := s.hmd.GetDecentralizedRoom(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, scopeNotFound("update decentralized room")
	}
	if err := requireDecentralizedCommunityAccess(ctx, scope, room.DecentralizedID, "update decentralized room"); err != nil {
		return nil, err
	}
	result, err := s.hmd.UpdateDecentralizedRoom(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *decentralizedRoomService) UpdateDecentralizedRoomStatus(ctx context.Context, input UpdateDecentralizedRoomStatusInput) (*hmdmodel.HmdRoomDecentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	room, err := s.hmd.GetDecentralizedRoom(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, scopeNotFound("update decentralized room status")
	}
	if err := requireDecentralizedCommunityAccess(ctx, scope, room.DecentralizedID, "update decentralized room status"); err != nil {
		return nil, err
	}
	result, err := s.hmd.UpdateDecentralizedRoomStatus(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}
