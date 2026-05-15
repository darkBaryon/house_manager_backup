package publish

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
)

type decentralizedCommunityService struct {
	hmd       decentralizedCommunityScopeDomain
	publisher mutationPublisher
	access    publishAccessService
}

func newDecentralizedCommunityService(hmd decentralizedCommunityScopeDomain, publisher mutationPublisher, access publishAccessService) *decentralizedCommunityService {
	return &decentralizedCommunityService{hmd: hmd, publisher: publisher, access: access}
}

func (s *decentralizedCommunityService) CreateDecentralizedCommunity(ctx context.Context, input CreateDecentralizedCommunityInput) (*hmdmodel.HmdDecentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	result, err := s.hmd.CreateDecentralizedCommunity(ctx, input)
	createdID := mutationEntityID(result)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	if err != nil {
		return nil, rollbackCreateFailure(ctx, s.hmd.RollbackDecentralizedCommunityCreate, createdID, "create decentralized community", err)
	}
	if entity != nil {
		if err := registerRootScope(ctx, s.access, hpdmodel.HpdRootScopeTypeDecentralizedCommunity, entity.ID, scope.principal); err != nil {
			return nil, rollbackCreateFailure(ctx, s.hmd.RollbackDecentralizedCommunityCreate, entity.ID, "create decentralized community", err)
		}
	}
	return entity, nil
}

func (s *decentralizedCommunityService) GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdDecentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	community, err := s.hmd.GetDecentralizedCommunity(ctx, id)
	if err != nil {
		return nil, err
	}
	if community != nil {
		if err := requireDecentralizedCommunityAccess(ctx, scope, community.ID, "get decentralized community"); err != nil {
			return nil, err
		}
	}
	return community, nil
}

func (s *decentralizedCommunityService) ListDecentralizedCommunities(ctx context.Context, input ListDecentralizedCommunitiesInput) ([]hmdmodel.HmdDecentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	communities, err := s.hmd.ListDecentralizedCommunities(ctx, input)
	if err != nil {
		return nil, err
	}
	return filterDecentralizedCommunitiesByScope(ctx, scope, communities)
}

func (s *decentralizedCommunityService) UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*hmdmodel.HmdDecentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	if err := requireDecentralizedCommunityAccess(ctx, scope, input.ID, "update decentralized community"); err != nil {
		return nil, err
	}
	result, err := s.hmd.UpdateDecentralizedCommunity(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}
