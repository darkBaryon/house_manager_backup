package publish

import (
	"context"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type decentralizedCommunityService struct {
	hmd       decentralizedCommunityDomain
	publisher mutationPublisher
}

func newDecentralizedCommunityService(hmd decentralizedCommunityDomain, publisher mutationPublisher) *decentralizedCommunityService {
	return &decentralizedCommunityService{hmd: hmd, publisher: publisher}
}

func (s *decentralizedCommunityService) CreateDecentralizedCommunity(ctx context.Context, input CreateDecentralizedCommunityInput) (*model.HmdDecentralized, error) {
	result, err := s.hmd.CreateDecentralizedCommunity(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *decentralizedCommunityService) GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*model.HmdDecentralized, error) {
	return s.hmd.GetDecentralizedCommunity(ctx, id)
}

func (s *decentralizedCommunityService) ListDecentralizedCommunities(ctx context.Context, input ListDecentralizedCommunitiesInput) ([]model.HmdDecentralized, error) {
	return s.hmd.ListDecentralizedCommunities(ctx, input)
}

func (s *decentralizedCommunityService) UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*model.HmdDecentralized, error) {
	result, err := s.hmd.UpdateDecentralizedCommunity(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}
