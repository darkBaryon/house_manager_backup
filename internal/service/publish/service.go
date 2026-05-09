package publish

import (
	"context"

	"house-manager/internal/model"
	hmdsvc "house-manager/internal/service/publish/hmd"
	hpdsvc "house-manager/internal/service/publish/hpd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// PublishService 是发房域的业务入口，handler 只依赖这一层。
type PublishService struct {
	hmd hmdService
	hpd hpdApplier
}

func NewPublishService(hmd *hmdsvc.Service, hpd *hpdsvc.Service) *PublishService {
	return newPublishService(hmd, hpd)
}

func newPublishService(hmd hmdService, hpd hpdApplier) *PublishService {
	return &PublishService{hmd: hmd, hpd: hpd}
}

type hmdService interface {
	CreateCentralizedProject(ctx context.Context, input CreateCentralizedProjectInput) (*hmdsvc.HmdMutationResult[model.HmdCentralized], error)
	GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error)
	ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]model.HmdCentralized, error)
	UpdateCentralizedProject(ctx context.Context, input UpdateCentralizedProjectInput) (*hmdsvc.HmdMutationResult[model.HmdCentralized], error)
	CreateBuilding(ctx context.Context, input CreateBuildingInput) (*hmdsvc.HmdMutationResult[model.HmdBuilding], error)
	GetBuilding(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error)
	ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error)
	UpdateBuilding(ctx context.Context, input UpdateBuildingInput) (*hmdsvc.HmdMutationResult[model.HmdBuilding], error)
	CreateRoomType(ctx context.Context, input CreateRoomTypeInput) (*hmdsvc.HmdMutationResult[model.HmdRoomTypeCentralized], error)
	GetRoomType(ctx context.Context, id bson.ObjectID) (*model.HmdRoomTypeCentralized, error)
	ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error)
	ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error)
	UpdateRoomType(ctx context.Context, input UpdateRoomTypeInput) (*hmdsvc.HmdMutationResult[model.HmdRoomTypeCentralized], error)
	CreateCentralizedRoom(ctx context.Context, input CreateCentralizedRoomInput) (*hmdsvc.HmdMutationResult[model.HmdRoomCentralized], error)
	GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error)
	ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error)
	ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error)
	UpdateCentralizedRoom(ctx context.Context, input UpdateCentralizedRoomInput) (*hmdsvc.HmdMutationResult[model.HmdRoomCentralized], error)
	UpdateCentralizedRoomStatus(ctx context.Context, input UpdateCentralizedRoomStatusInput) (*hmdsvc.HmdMutationResult[model.HmdRoomCentralized], error)
	CreateDecentralizedCommunity(ctx context.Context, input CreateDecentralizedCommunityInput) (*hmdsvc.HmdMutationResult[model.HmdDecentralized], error)
	GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*model.HmdDecentralized, error)
	ListDecentralizedCommunities(ctx context.Context, input ListDecentralizedCommunitiesInput) ([]model.HmdDecentralized, error)
	UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*hmdsvc.HmdMutationResult[model.HmdDecentralized], error)
	CreateDecentralizedRoom(ctx context.Context, input CreateDecentralizedRoomInput) (*hmdsvc.HmdMutationResult[model.HmdRoomDecentralized], error)
	GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomDecentralized, error)
	ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]model.HmdRoomDecentralized, error)
	UpdateDecentralizedRoom(ctx context.Context, input UpdateDecentralizedRoomInput) (*hmdsvc.HmdMutationResult[model.HmdRoomDecentralized], error)
	UpdateDecentralizedRoomStatus(ctx context.Context, input UpdateDecentralizedRoomStatusInput) (*hmdsvc.HmdMutationResult[model.HmdRoomDecentralized], error)
}

type hpdApplier interface {
	Apply(ctx context.Context, changes []hmdsvc.HmdChange) error
}

func (s *PublishService) CreateCentralizedProject(ctx context.Context, input CreateCentralizedProjectInput) (*model.HmdCentralized, error) {
	result, err := s.hmd.CreateCentralizedProject(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error) {
	return s.hmd.GetCentralizedProject(ctx, id)
}

func (s *PublishService) ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]model.HmdCentralized, error) {
	return s.hmd.ListCentralizedProjects(ctx, input)
}

func (s *PublishService) UpdateCentralizedProject(ctx context.Context, input UpdateCentralizedProjectInput) (*model.HmdCentralized, error) {
	result, err := s.hmd.UpdateCentralizedProject(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) CreateBuilding(ctx context.Context, input CreateBuildingInput) (*model.HmdBuilding, error) {
	result, err := s.hmd.CreateBuilding(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) GetBuilding(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error) {
	return s.hmd.GetBuilding(ctx, id)
}

func (s *PublishService) ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error) {
	return s.hmd.ListBuildingsByProject(ctx, projectID)
}

func (s *PublishService) UpdateBuilding(ctx context.Context, input UpdateBuildingInput) (*model.HmdBuilding, error) {
	result, err := s.hmd.UpdateBuilding(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) CreateRoomType(ctx context.Context, input CreateRoomTypeInput) (*model.HmdRoomTypeCentralized, error) {
	result, err := s.hmd.CreateRoomType(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) GetRoomType(ctx context.Context, id bson.ObjectID) (*model.HmdRoomTypeCentralized, error) {
	return s.hmd.GetRoomType(ctx, id)
}

func (s *PublishService) ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error) {
	return s.hmd.ListRoomTypesByProject(ctx, projectID)
}

func (s *PublishService) ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error) {
	return s.hmd.ListRoomTypesByBuilding(ctx, buildingID)
}

func (s *PublishService) UpdateRoomType(ctx context.Context, input UpdateRoomTypeInput) (*model.HmdRoomTypeCentralized, error) {
	result, err := s.hmd.UpdateRoomType(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) CreateCentralizedRoom(ctx context.Context, input CreateCentralizedRoomInput) (*model.HmdRoomCentralized, error) {
	result, err := s.hmd.CreateCentralizedRoom(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error) {
	return s.hmd.GetCentralizedRoom(ctx, id)
}

func (s *PublishService) ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	return s.hmd.ListCentralizedRoomsByProject(ctx, projectID)
}

func (s *PublishService) ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error) {
	return s.hmd.ListCentralizedRoomsByBuilding(ctx, buildingID)
}

func (s *PublishService) UpdateCentralizedRoom(ctx context.Context, input UpdateCentralizedRoomInput) (*model.HmdRoomCentralized, error) {
	result, err := s.hmd.UpdateCentralizedRoom(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) UpdateCentralizedRoomStatus(ctx context.Context, input UpdateCentralizedRoomStatusInput) (*model.HmdRoomCentralized, error) {
	result, err := s.hmd.UpdateCentralizedRoomStatus(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) CreateDecentralizedCommunity(ctx context.Context, input CreateDecentralizedCommunityInput) (*model.HmdDecentralized, error) {
	result, err := s.hmd.CreateDecentralizedCommunity(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*model.HmdDecentralized, error) {
	return s.hmd.GetDecentralizedCommunity(ctx, id)
}

func (s *PublishService) ListDecentralizedCommunities(ctx context.Context, input ListDecentralizedCommunitiesInput) ([]model.HmdDecentralized, error) {
	return s.hmd.ListDecentralizedCommunities(ctx, input)
}

func (s *PublishService) UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*model.HmdDecentralized, error) {
	result, err := s.hmd.UpdateDecentralizedCommunity(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) CreateDecentralizedRoom(ctx context.Context, input CreateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error) {
	result, err := s.hmd.CreateDecentralizedRoom(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomDecentralized, error) {
	return s.hmd.GetDecentralizedRoom(ctx, id)
}

func (s *PublishService) ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]model.HmdRoomDecentralized, error) {
	return s.hmd.ListDecentralizedRoomsByCommunity(ctx, decentralizedID)
}

func (s *PublishService) UpdateDecentralizedRoom(ctx context.Context, input UpdateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error) {
	result, err := s.hmd.UpdateDecentralizedRoom(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func (s *PublishService) UpdateDecentralizedRoomStatus(ctx context.Context, input UpdateDecentralizedRoomStatusInput) (*model.HmdRoomDecentralized, error) {
	result, err := s.hmd.UpdateDecentralizedRoomStatus(ctx, input)
	return resolveHmdMutation(ctx, s.hpd, result, err)
}

func resolveHmdMutation[T any](ctx context.Context, hpd hpdApplier, result *hmdsvc.HmdMutationResult[T], err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	if err := hpd.Apply(ctx, result.Changes); err != nil {
		return nil, err
	}
	return result.Entity, nil
}
