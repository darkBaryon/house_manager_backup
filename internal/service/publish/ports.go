package publish

import (
	"context"

	hmddomain "house-manager/internal/domain/hmd"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type centralizedProjectDomain interface {
	CreateCentralizedProject(ctx context.Context, input CreateCentralizedProjectInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdCentralized], error)
	RollbackCentralizedProjectCreate(ctx context.Context, id bson.ObjectID) error
	GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdCentralized, error)
	ListCentralizedProjects(ctx context.Context, input ListCentralizedProjectsInput) ([]hmdmodel.HmdCentralized, error)
	ListCentralizedProjectsByIDs(ctx context.Context, ids []bson.ObjectID, input ListCentralizedProjectsInput) ([]hmdmodel.HmdCentralized, error)
	UpdateCentralizedProject(ctx context.Context, input UpdateCentralizedProjectInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdCentralized], error)
}

type centralizedProjectScopeDomain interface {
	centralizedProjectDomain
	GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomCentralized, error)
}

type buildingDomain interface {
	CreateBuilding(ctx context.Context, input CreateBuildingInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdBuilding], error)
	GetBuilding(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdBuilding, error)
	ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdBuilding, error)
	UpdateBuilding(ctx context.Context, input UpdateBuildingInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdBuilding], error)
}

type buildingScopeDomain interface {
	buildingDomain
	GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomCentralized, error)
}

type roomTypeDomain interface {
	CreateRoomType(ctx context.Context, input CreateRoomTypeInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomTypeCentralized], error)
	GetRoomType(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomTypeCentralized, error)
	ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomTypeCentralized, error)
	ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomTypeCentralized, error)
	UpdateRoomType(ctx context.Context, input UpdateRoomTypeInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomTypeCentralized], error)
}

type roomTypeScopeDomain interface {
	roomTypeDomain
	GetBuilding(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdBuilding, error)
	GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomCentralized, error)
}

type centralizedRoomDomain interface {
	CreateCentralizedRoom(ctx context.Context, input CreateCentralizedRoomInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized], error)
	GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomCentralized, error)
	ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error)
	ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomCentralized, error)
	UpdateCentralizedRoom(ctx context.Context, input UpdateCentralizedRoomInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized], error)
	UpdateCentralizedRoomStatus(ctx context.Context, input UpdateCentralizedRoomStatusInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomCentralized], error)
}

type centralizedRoomScopeDomain interface {
	centralizedRoomDomain
	GetBuilding(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdBuilding, error)
	GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomCentralized, error)
	GetRoomType(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomTypeCentralized, error)
}

type decentralizedCommunityDomain interface {
	CreateDecentralizedCommunity(ctx context.Context, input CreateDecentralizedCommunityInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdDecentralized], error)
	RollbackDecentralizedCommunityCreate(ctx context.Context, id bson.ObjectID) error
	GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdDecentralized, error)
	ListDecentralizedCommunities(ctx context.Context, input ListDecentralizedCommunitiesInput) ([]hmdmodel.HmdDecentralized, error)
	ListDecentralizedCommunitiesByIDs(ctx context.Context, ids []bson.ObjectID, input ListDecentralizedCommunitiesInput) ([]hmdmodel.HmdDecentralized, error)
	UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdDecentralized], error)
}

type decentralizedCommunityScopeDomain interface {
	decentralizedCommunityDomain
	GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomDecentralized, error)
}

type decentralizedRoomDomain interface {
	CreateDecentralizedRoom(ctx context.Context, input CreateDecentralizedRoomInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomDecentralized], error)
	GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomDecentralized, error)
	ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]hmdmodel.HmdRoomDecentralized, error)
	UpdateDecentralizedRoom(ctx context.Context, input UpdateDecentralizedRoomInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomDecentralized], error)
	UpdateDecentralizedRoomStatus(ctx context.Context, input UpdateDecentralizedRoomStatusInput) (*hmddomain.HmdMutationResult[hmdmodel.HmdRoomDecentralized], error)
}

type listingProjectionApplier interface {
	Apply(ctx context.Context, changes []hmddomain.HmdChange) error
}

type publishRootScopeRegistrar interface {
	UpsertRootScopeForPrincipal(ctx context.Context, rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, principal session.Principal) (*hpdmodel.HpdRootScopeRelation, error)
}

type publishAccessService interface {
	publishRootScopeRegistrar
	ListAccessibleProjectIDs(ctx context.Context, principal session.Principal) ([]bson.ObjectID, error)
	ListAccessibleCommunityIDs(ctx context.Context, principal session.Principal) ([]bson.ObjectID, error)
	CanAccessProjectForPrincipal(ctx context.Context, projectID bson.ObjectID, principal session.Principal) (bool, error)
	CanAccessCommunityForPrincipal(ctx context.Context, communityID bson.ObjectID, principal session.Principal) (bool, error)
}
