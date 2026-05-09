package publish

import (
	"context"

	"house-manager/internal/model"
	hmdsvc "house-manager/internal/service/hmd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type centralizedProjectService interface {
	CreateCentralizedProject(ctx context.Context, input CreateCentralizedProjectInput) (*hmdsvc.HmdMutationResult[model.HmdCentralized], error)
	GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error)
	ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]model.HmdCentralized, error)
	UpdateCentralizedProject(ctx context.Context, input UpdateCentralizedProjectInput) (*hmdsvc.HmdMutationResult[model.HmdCentralized], error)
}

type buildingService interface {
	CreateBuilding(ctx context.Context, input CreateBuildingInput) (*hmdsvc.HmdMutationResult[model.HmdBuilding], error)
	GetBuilding(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error)
	ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error)
	UpdateBuilding(ctx context.Context, input UpdateBuildingInput) (*hmdsvc.HmdMutationResult[model.HmdBuilding], error)
}

type roomTypeService interface {
	CreateRoomType(ctx context.Context, input CreateRoomTypeInput) (*hmdsvc.HmdMutationResult[model.HmdRoomTypeCentralized], error)
	GetRoomType(ctx context.Context, id bson.ObjectID) (*model.HmdRoomTypeCentralized, error)
	ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error)
	ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error)
	UpdateRoomType(ctx context.Context, input UpdateRoomTypeInput) (*hmdsvc.HmdMutationResult[model.HmdRoomTypeCentralized], error)
}

type centralizedRoomService interface {
	CreateCentralizedRoom(ctx context.Context, input CreateCentralizedRoomInput) (*hmdsvc.HmdMutationResult[model.HmdRoomCentralized], error)
	GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error)
	ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error)
	ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error)
	UpdateCentralizedRoom(ctx context.Context, input UpdateCentralizedRoomInput) (*hmdsvc.HmdMutationResult[model.HmdRoomCentralized], error)
	UpdateCentralizedRoomStatus(ctx context.Context, input UpdateCentralizedRoomStatusInput) (*hmdsvc.HmdMutationResult[model.HmdRoomCentralized], error)
}

type decentralizedCommunityService interface {
	CreateDecentralizedCommunity(ctx context.Context, input CreateDecentralizedCommunityInput) (*hmdsvc.HmdMutationResult[model.HmdDecentralized], error)
	GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*model.HmdDecentralized, error)
	ListDecentralizedCommunities(ctx context.Context, input ListDecentralizedCommunitiesInput) ([]model.HmdDecentralized, error)
	UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*hmdsvc.HmdMutationResult[model.HmdDecentralized], error)
}

type decentralizedRoomService interface {
	CreateDecentralizedRoom(ctx context.Context, input CreateDecentralizedRoomInput) (*hmdsvc.HmdMutationResult[model.HmdRoomDecentralized], error)
	GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomDecentralized, error)
	ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]model.HmdRoomDecentralized, error)
	UpdateDecentralizedRoom(ctx context.Context, input UpdateDecentralizedRoomInput) (*hmdsvc.HmdMutationResult[model.HmdRoomDecentralized], error)
	UpdateDecentralizedRoomStatus(ctx context.Context, input UpdateDecentralizedRoomStatusInput) (*hmdsvc.HmdMutationResult[model.HmdRoomDecentralized], error)
}

type hpdApplier interface {
	Apply(ctx context.Context, changes []hmdsvc.HmdChange) error
}
