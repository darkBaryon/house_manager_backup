package publish

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
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
	if _, err := newPublishScope(ctx, s.access); err != nil {
		return nil, err
	}
	result, err := s.hmd.CreateDecentralizedCommunity(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
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
		if err := requireDecentralizedCommunityAccess(ctx, scope, s.hmd, community.ID, "get decentralized community"); err != nil {
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
	return filterDecentralizedCommunitiesByScope(ctx, scope, s.hmd, communities)
}

func (s *decentralizedCommunityService) UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*hmdmodel.HmdDecentralized, error) {
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		return nil, err
	}
	if err := scope.requireGlobal("update decentralized community"); err != nil {
		return nil, err
	}
	result, err := s.hmd.UpdateDecentralizedCommunity(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}
