package publish

import (
	"context"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *PublishService) CreateDecentralizedCommunity(ctx context.Context, input CreateDecentralizedCommunityInput) (*model.HmdDecentralized, error) {
	result, err := s.decentralizedCommunities.CreateDecentralizedCommunity(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*model.HmdDecentralized, error) {
	return s.decentralizedCommunities.GetDecentralizedCommunity(ctx, id)
}

func (s *PublishService) ListDecentralizedCommunities(ctx context.Context, input ListDecentralizedCommunitiesInput) ([]model.HmdDecentralized, error) {
	return s.decentralizedCommunities.ListDecentralizedCommunities(ctx, input)
}

func (s *PublishService) UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*model.HmdDecentralized, error) {
	result, err := s.decentralizedCommunities.UpdateDecentralizedCommunity(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}
