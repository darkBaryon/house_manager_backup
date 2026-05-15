package publish

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
)

type centralizedRoomReader interface {
	GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomCentralized, error)
}

type decentralizedRoomReader interface {
	GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomDecentralized, error)
}

func filterCentralizedProjectsByScope(ctx context.Context, scope PublishScope, hmd centralizedRoomReader, projects []hmdmodel.HmdCentralized) ([]hmdmodel.HmdCentralized, error) {
	if scope.IsGlobal() {
		return projects, nil
	}
	projectIDs, err := centralizedProjectIDSetForScope(ctx, scope, hmd)
	if err != nil {
		return nil, err
	}
	filtered := make([]hmdmodel.HmdCentralized, 0, len(projects))
	for _, project := range projects {
		if _, ok := projectIDs[project.ID]; ok {
			filtered = append(filtered, project)
		}
	}
	return filtered, nil
}

func filterDecentralizedCommunitiesByScope(ctx context.Context, scope PublishScope, hmd decentralizedRoomReader, communities []hmdmodel.HmdDecentralized) ([]hmdmodel.HmdDecentralized, error) {
	if scope.IsGlobal() {
		return communities, nil
	}
	communityIDs, err := decentralizedCommunityIDSetForScope(ctx, scope, hmd)
	if err != nil {
		return nil, err
	}
	filtered := make([]hmdmodel.HmdDecentralized, 0, len(communities))
	for _, community := range communities {
		if _, ok := communityIDs[community.ID]; ok {
			filtered = append(filtered, community)
		}
	}
	return filtered, nil
}

func filterCentralizedRoomsByScope(ctx context.Context, scope PublishScope, rooms []hmdmodel.HmdRoomCentralized) ([]hmdmodel.HmdRoomCentralized, error) {
	if scope.IsGlobal() {
		return rooms, nil
	}
	roomIDs, err := scope.accessibleSourceIDSet(ctx, hpdmodel.HpdSourceTypeCentralizedRoom)
	if err != nil {
		return nil, err
	}
	filtered := make([]hmdmodel.HmdRoomCentralized, 0, len(rooms))
	for _, room := range rooms {
		if _, ok := roomIDs[room.ID]; ok {
			filtered = append(filtered, room)
		}
	}
	return filtered, nil
}

func filterDecentralizedRoomsByScope(ctx context.Context, scope PublishScope, rooms []hmdmodel.HmdRoomDecentralized) ([]hmdmodel.HmdRoomDecentralized, error) {
	if scope.IsGlobal() {
		return rooms, nil
	}
	roomIDs, err := scope.accessibleSourceIDSet(ctx, hpdmodel.HpdSourceTypeDecentralizedRoom)
	if err != nil {
		return nil, err
	}
	filtered := make([]hmdmodel.HmdRoomDecentralized, 0, len(rooms))
	for _, room := range rooms {
		if _, ok := roomIDs[room.ID]; ok {
			filtered = append(filtered, room)
		}
	}
	return filtered, nil
}

func canAccessCentralizedProject(ctx context.Context, scope PublishScope, hmd centralizedRoomReader, projectID bson.ObjectID) (bool, error) {
	if scope.IsGlobal() {
		return true, nil
	}
	projectIDs, err := centralizedProjectIDSetForScope(ctx, scope, hmd)
	if err != nil {
		return false, err
	}
	_, ok := projectIDs[projectID]
	return ok, nil
}

func requireCentralizedProjectAccess(ctx context.Context, scope PublishScope, hmd centralizedRoomReader, projectID bson.ObjectID, action string) error {
	allowed, err := canAccessCentralizedProject(ctx, scope, hmd, projectID)
	if err != nil {
		return err
	}
	if !allowed {
		return scopeNotFound(action)
	}
	return nil
}

func canAccessDecentralizedCommunity(ctx context.Context, scope PublishScope, hmd decentralizedRoomReader, decentralizedID bson.ObjectID) (bool, error) {
	if scope.IsGlobal() {
		return true, nil
	}
	communityIDs, err := decentralizedCommunityIDSetForScope(ctx, scope, hmd)
	if err != nil {
		return false, err
	}
	_, ok := communityIDs[decentralizedID]
	return ok, nil
}

func requireDecentralizedCommunityAccess(ctx context.Context, scope PublishScope, hmd decentralizedRoomReader, decentralizedID bson.ObjectID, action string) error {
	allowed, err := canAccessDecentralizedCommunity(ctx, scope, hmd, decentralizedID)
	if err != nil {
		return err
	}
	if !allowed {
		return scopeNotFound(action)
	}
	return nil
}

func centralizedProjectIDSetForScope(ctx context.Context, scope PublishScope, hmd centralizedRoomReader) (map[bson.ObjectID]struct{}, error) {
	roomIDs, err := scope.accessibleSourceIDSet(ctx, hpdmodel.HpdSourceTypeCentralizedRoom)
	if err != nil {
		return nil, err
	}
	projectIDs := make(map[bson.ObjectID]struct{}, len(roomIDs))
	for roomID := range roomIDs {
		room, err := hmd.GetCentralizedRoom(ctx, roomID)
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return nil, err
		}
		if room == nil || room.ProjectID.IsZero() {
			continue
		}
		projectIDs[room.ProjectID] = struct{}{}
	}
	return projectIDs, nil
}

func decentralizedCommunityIDSetForScope(ctx context.Context, scope PublishScope, hmd decentralizedRoomReader) (map[bson.ObjectID]struct{}, error) {
	roomIDs, err := scope.accessibleSourceIDSet(ctx, hpdmodel.HpdSourceTypeDecentralizedRoom)
	if err != nil {
		return nil, err
	}
	communityIDs := make(map[bson.ObjectID]struct{}, len(roomIDs))
	for roomID := range roomIDs {
		room, err := hmd.GetDecentralizedRoom(ctx, roomID)
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return nil, err
		}
		if room == nil || room.DecentralizedID.IsZero() {
			continue
		}
		communityIDs[room.DecentralizedID] = struct{}{}
	}
	return communityIDs, nil
}
