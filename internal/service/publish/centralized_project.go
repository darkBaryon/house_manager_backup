package publish

import (
	"context"

	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type centralizedProjectService struct {
	hmd       centralizedProjectDomain
	publisher mutationPublisher
}

func newCentralizedProjectService(hmd centralizedProjectDomain, publisher mutationPublisher) *centralizedProjectService {
	return &centralizedProjectService{hmd: hmd, publisher: publisher}
}

func (s *centralizedProjectService) CreateCentralizedProject(ctx context.Context, input CreateCentralizedProjectInput) (*model.HmdCentralized, error) {
	result, err := s.hmd.CreateCentralizedProject(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}

func (s *centralizedProjectService) GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error) {
	return s.hmd.GetCentralizedProject(ctx, id)
}

func (s *centralizedProjectService) ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]model.HmdCentralized, error) {
	return s.hmd.ListCentralizedProjects(ctx, input)
}

func (s *centralizedProjectService) UpdateCentralizedProject(ctx context.Context, input UpdateCentralizedProjectInput) (*model.HmdCentralized, error) {
	result, err := s.hmd.UpdateCentralizedProject(ctx, input)
	return resolveHmdMutation(ctx, s.publisher, result, err)
}
