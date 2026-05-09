package publish

import (
	"context"

	"house-manager/internal/handler"
	"house-manager/internal/model"
	publishsvc "house-manager/internal/service/publish"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PublishHandler struct {
	service publishService
}

func NewPublishHandler(service *publishsvc.PublishService) *PublishHandler {
	return newPublishHandler(service)
}

var _ handler.RouteRegistrar = (*PublishHandler)(nil)

type publishService interface {
	CreateCentralizedProject(ctx context.Context, input publishsvc.CreateCentralizedProjectInput) (*model.HmdCentralized, error)
	GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error)
	ListCentralizedProjects(ctx context.Context, input publishsvc.ListCentralizedProjectsInput) ([]model.HmdCentralized, error)
	UpdateCentralizedProject(ctx context.Context, input publishsvc.UpdateCentralizedProjectInput) (*model.HmdCentralized, error)
	CreateBuilding(ctx context.Context, input publishsvc.CreateBuildingInput) (*model.HmdBuilding, error)
	GetBuilding(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error)
	ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error)
	UpdateBuilding(ctx context.Context, input publishsvc.UpdateBuildingInput) (*model.HmdBuilding, error)
	CreateRoomType(ctx context.Context, input publishsvc.CreateRoomTypeInput) (*model.HmdRoomTypeCentralized, error)
	GetRoomType(ctx context.Context, id bson.ObjectID) (*model.HmdRoomTypeCentralized, error)
	ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error)
	ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomTypeCentralized, error)
	UpdateRoomType(ctx context.Context, input publishsvc.UpdateRoomTypeInput) (*model.HmdRoomTypeCentralized, error)
	CreateCentralizedRoom(ctx context.Context, input publishsvc.CreateCentralizedRoomInput) (*model.HmdRoomCentralized, error)
	GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error)
	ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error)
	ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error)
	UpdateCentralizedRoom(ctx context.Context, input publishsvc.UpdateCentralizedRoomInput) (*model.HmdRoomCentralized, error)
	UpdateCentralizedRoomStatus(ctx context.Context, input publishsvc.UpdateCentralizedRoomStatusInput) (*model.HmdRoomCentralized, error)
	CreateDecentralizedCommunity(ctx context.Context, input publishsvc.CreateDecentralizedCommunityInput) (*model.HmdDecentralized, error)
	GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*model.HmdDecentralized, error)
	ListDecentralizedCommunities(ctx context.Context, input publishsvc.ListDecentralizedCommunitiesInput) ([]model.HmdDecentralized, error)
	UpdateDecentralizedCommunity(ctx context.Context, input publishsvc.UpdateDecentralizedCommunityInput) (*model.HmdDecentralized, error)
	CreateDecentralizedRoom(ctx context.Context, input publishsvc.CreateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error)
	GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomDecentralized, error)
	ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]model.HmdRoomDecentralized, error)
	UpdateDecentralizedRoom(ctx context.Context, input publishsvc.UpdateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error)
	UpdateDecentralizedRoomStatus(ctx context.Context, input publishsvc.UpdateDecentralizedRoomStatusInput) (*model.HmdRoomDecentralized, error)
}

var _ publishService = (*publishsvc.PublishService)(nil)

func newPublishHandler(service publishService) *PublishHandler {
	return &PublishHandler{service: service}
}
