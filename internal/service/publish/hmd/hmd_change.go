package hmd

import "go.mongodb.org/mongo-driver/v2/bson"

type HmdChangeAction string

const (
	HmdChangeCreated       HmdChangeAction = "created"
	HmdChangeUpdated       HmdChangeAction = "updated"
	HmdChangeStatusUpdated HmdChangeAction = "status_updated"
)

type HmdEntityType string

const (
	HmdEntityCentralizedProject     HmdEntityType = "centralized_project"
	HmdEntityBuilding               HmdEntityType = "building"
	HmdEntityRoomTypeCentralized    HmdEntityType = "room_type_centralized"
	HmdEntityRoomCentralized        HmdEntityType = "room_centralized"
	HmdEntityDecentralizedCommunity HmdEntityType = "decentralized_community"
	HmdEntityRoomDecentralized      HmdEntityType = "room_decentralized"
)

type HmdProjectionScope string

const (
	HmdScopeCentralizedProject     HmdProjectionScope = "centralized_project"
	HmdScopeBuilding               HmdProjectionScope = "building"
	HmdScopeRoomTypeCentralized    HmdProjectionScope = "room_type_centralized"
	HmdScopeCentralizedRoom        HmdProjectionScope = "centralized_room"
	HmdScopeDecentralizedCommunity HmdProjectionScope = "decentralized_community"
	HmdScopeDecentralizedRoom      HmdProjectionScope = "decentralized_room"
)

type HmdChange struct {
	Action          HmdChangeAction    `json:"action"`
	EntityType      HmdEntityType      `json:"entityType"`
	EntityID        bson.ObjectID      `json:"entityId"`
	Scope           HmdProjectionScope `json:"scope"`
	ProjectID       bson.ObjectID      `json:"projectId,omitempty"`
	BuildingID      bson.ObjectID      `json:"buildingId,omitempty"`
	RoomTypeID      bson.ObjectID      `json:"roomTypeId,omitempty"`
	DecentralizedID bson.ObjectID      `json:"decentralizedId,omitempty"`
}

type HmdMutationResult[T any] struct {
	Entity  *T
	Changes []HmdChange
}

func hmdMutationResult[T any](entity *T, changes ...HmdChange) *HmdMutationResult[T] {
	return &HmdMutationResult[T]{
		Entity:  entity,
		Changes: changes,
	}
}
