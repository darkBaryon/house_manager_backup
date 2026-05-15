package publish

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
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
	if err := requireDecentralizedCommunityAccess(ctx, scope, s.hmd, input.DecentralizedID, "create decentralized room"); err != nil {
		return nil, err
	}
	result, err := s.hmd.CreateDecentralizedRoom(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	if err != nil {
		return nil, err
	}
	if entity != nil {
		if err := registerRoomEntrust(ctx, s.access, hpdmodel.HpdSourceTypeDecentralizedRoom, entity.ID, scope.principal); err != nil {
			return nil, err
		}
	}
	return entity, nil
}

func (s *decentralizedRoomService) GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomDecentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	if err := scope.requireCanAccessSource(ctx, hpdmodel.HpdSourceTypeDecentralizedRoom, id, "get decentralized room"); err != nil {
		return nil, err
	}
	return s.hmd.GetDecentralizedRoom(ctx, id)
}

func (s *decentralizedRoomService) ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]hmdmodel.HmdRoomDecentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	rooms, err := s.hmd.ListDecentralizedRoomsByCommunity(ctx, decentralizedID)
	if err != nil {
		return nil, err
	}
	return filterDecentralizedRoomsByScope(ctx, scope, rooms)
}

func (s *decentralizedRoomService) UpdateDecentralizedRoom(ctx context.Context, input UpdateDecentralizedRoomInput) (*hmdmodel.HmdRoomDecentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	if err := scope.requireCanAccessSource(ctx, hpdmodel.HpdSourceTypeDecentralizedRoom, input.ID, "update decentralized room"); err != nil {
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
	if err := scope.requireCanAccessSource(ctx, hpdmodel.HpdSourceTypeDecentralizedRoom, input.ID, "update decentralized room status"); err != nil {
		return nil, err
	}
	result, err := s.hmd.UpdateDecentralizedRoomStatus(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}
