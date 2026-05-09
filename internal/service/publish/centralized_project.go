package publish

import (
	"context"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *PublishService) CreateCentralizedProject(ctx context.Context, input CreateCentralizedProjectInput) (*model.HmdCentralized, error) {
	result, err := s.centralizedProjects.CreateCentralizedProject(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error) {
	return s.centralizedProjects.GetCentralizedProject(ctx, id)
}

func (s *PublishService) ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]model.HmdCentralized, error) {
	return s.centralizedProjects.ListCentralizedProjects(ctx, input)
}

func (s *PublishService) UpdateCentralizedProject(ctx context.Context, input UpdateCentralizedProjectInput) (*model.HmdCentralized, error) {
	result, err := s.centralizedProjects.UpdateCentralizedProject(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}
