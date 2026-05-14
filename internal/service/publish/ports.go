package publish

import (
	"context"

	hmddomain "house-manager/internal/domain/hmd"
	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type centralizedProjectDomain interface {
	CreateCentralizedProject(ctx context.Context, input CreateCentralizedProjectInput) (*hmddomain.HmdMutationResult[model.HmdCentralized], error)
	GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error)
	ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]model.HmdCentralized, error)
	UpdateCentralizedProject(ctx context.Context, input UpdateCentralizedProjectInput) (*hmddomain.HmdMutationResult[model.HmdCentralized], error)
}

type buildingDomain interface {
	CreateBuilding(ctx context.Context, input CreateBuildingInput) (*hmddomain.HmdMutationResult[model.HmdBuilding], error)
	GetBuilding(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error)
	ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error)
	UpdateBuilding(ctx context.Context, input UpdateBuildingInput) (*hmddomain.HmdMutationResult[model.HmdBuilding], error)
}

type roomTypeDomain interface {
	CreateRoomType(ctx context.Context, input CreateRoomTypeInput) (*hmddomain.HmdMutationResult[model.HmdRoomTypeCentralized], error)
	GetRoomType(ctx context.Context, id bson.ObjectID) (*model.HmdRoomTypeCentralized, error)
	ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error)
	ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error)
	UpdateRoomType(ctx context.Context, input UpdateRoomTypeInput) (*hmddomain.HmdMutationResult[model.HmdRoomTypeCentralized], error)
}

type centralizedRoomDomain interface {
	CreateCentralizedRoom(ctx context.Context, input CreateCentralizedRoomInput) (*hmddomain.HmdMutationResult[model.HmdRoomCentralized], error)
	GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error)
	ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error)
	ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error)
	UpdateCentralizedRoom(ctx context.Context, input UpdateCentralizedRoomInput) (*hmddomain.HmdMutationResult[model.HmdRoomCentralized], error)
	UpdateCentralizedRoomStatus(ctx context.Context, input UpdateCentralizedRoomStatusInput) (*hmddomain.HmdMutationResult[model.HmdRoomCentralized], error)
}

type decentralizedCommunityDomain interface {
	CreateDecentralizedCommunity(ctx context.Context, input CreateDecentralizedCommunityInput) (*hmddomain.HmdMutationResult[model.HmdDecentralized], error)
	GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*model.HmdDecentralized, error)
	ListDecentralizedCommunities(ctx context.Context, input ListDecentralizedCommunitiesInput) ([]model.HmdDecentralized, error)
	UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*hmddomain.HmdMutationResult[model.HmdDecentralized], error)
}

type decentralizedRoomDomain interface {
	CreateDecentralizedRoom(ctx context.Context, input CreateDecentralizedRoomInput) (*hmddomain.HmdMutationResult[model.HmdRoomDecentralized], error)
	GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomDecentralized, error)
	ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]model.HmdRoomDecentralized, error)
	UpdateDecentralizedRoom(ctx context.Context, input UpdateDecentralizedRoomInput) (*hmddomain.HmdMutationResult[model.HmdRoomDecentralized], error)
	UpdateDecentralizedRoomStatus(ctx context.Context, input UpdateDecentralizedRoomStatusInput) (*hmddomain.HmdMutationResult[model.HmdRoomDecentralized], error)
}

type hpdApplier interface {
	Apply(ctx context.Context, changes []hmddomain.HmdChange) error
}
