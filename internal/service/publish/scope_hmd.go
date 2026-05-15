package publish

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
)

func filterCentralizedProjectsByScope(ctx context.Context, scope PublishScope, projects []hmdmodel.HmdCentralized) ([]hmdmodel.HmdCentralized, error) {
	projectIDs, err := scope.accessibleProjectIDSet(ctx)
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

func filterDecentralizedCommunitiesByScope(ctx context.Context, scope PublishScope, communities []hmdmodel.HmdDecentralized) ([]hmdmodel.HmdDecentralized, error) {
	communityIDs, err := scope.accessibleCommunityIDSet(ctx)
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

func canAccessCentralizedProject(ctx context.Context, scope PublishScope, projectID bson.ObjectID) (bool, error) {
	projectIDs, err := scope.accessibleProjectIDSet(ctx)
	if err != nil {
		return false, err
	}
	_, ok := projectIDs[projectID]
	return ok, nil
}

func requireCentralizedProjectAccess(ctx context.Context, scope PublishScope, projectID bson.ObjectID, action string) error {
	allowed, err := canAccessCentralizedProject(ctx, scope, projectID)
	if err != nil {
		return err
	}
	if !allowed {
		return scopeNotFound(action)
	}
	return nil
}

func canAccessDecentralizedCommunity(ctx context.Context, scope PublishScope, decentralizedID bson.ObjectID) (bool, error) {
	communityIDs, err := scope.accessibleCommunityIDSet(ctx)
	if err != nil {
		return false, err
	}
	_, ok := communityIDs[decentralizedID]
	return ok, nil
}

func requireDecentralizedCommunityAccess(ctx context.Context, scope PublishScope, decentralizedID bson.ObjectID, action string) error {
	allowed, err := canAccessDecentralizedCommunity(ctx, scope, decentralizedID)
	if err != nil {
		return err
	}
	if !allowed {
		return scopeNotFound(action)
	}
	return nil
}
