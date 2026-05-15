package publish

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
)

type decentralizedCommunityService struct {
	hmd       decentralizedCommunityScopeDomain
	publisher mutationPublisher
	access    publishAccessService
}

func newDecentralizedCommunityService(hmd decentralizedCommunityScopeDomain, publisher mutationPublisher, access publishAccessService) *decentralizedCommunityService {
	return &decentralizedCommunityService{hmd: hmd, publisher: publisher, access: access}
}

func (s *decentralizedCommunityService) CreateDecentralizedCommunity(ctx context.Context, input CreateDecentralizedCommunityInput) (*hmdmodel.HmdDecentralized, error) {
	logPublishInfo(ctx, "publish.community.create.start", "community_name", input.CommunityName, "city", input.City, "district", input.District)
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.community.create.success", "publish.community.create.failed", err, "community_name", input.CommunityName)
		return nil, err
	}
	result, err := s.hmd.CreateDecentralizedCommunity(ctx, input)
	createdID := mutationEntityID(result)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	if err != nil {
		err = rollbackCreateFailure(ctx, s.hmd.RollbackDecentralizedCommunityCreate, createdID, "publish.community.create", err)
		logPublishResult(ctx, "publish.community.create.success", "publish.community.create.failed", err, "community_id", createdID.Hex())
		return nil, err
	}
	if entity != nil {
		logPublishInfo(ctx, "publish.community.create.hmd_created", "community_id", entity.ID.Hex(), "community_name", entity.CommunityName)
		if err := registerRootScope(ctx, s.access, hpdmodel.HpdRootScopeTypeDecentralizedCommunity, entity.ID, scope.principal); err != nil {
			err = rollbackCreateFailure(ctx, s.hmd.RollbackDecentralizedCommunityCreate, entity.ID, "publish.community.create", err)
			logPublishResult(ctx, "publish.community.create.success", "publish.community.create.failed", err, "community_id", entity.ID.Hex())
			return nil, err
		}
		logPublishInfo(ctx, "publish.community.create.root_scope_created", "community_id", entity.ID.Hex())
	}
	logPublishInfo(ctx, "publish.community.create.success", "community_id", entity.ID.Hex(), "community_name", entity.CommunityName)
	return entity, nil
}

func (s *decentralizedCommunityService) GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdDecentralized, error) {
	logPublishInfo(ctx, "publish.community.detail.start", "community_id", id.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.community.detail.success", "publish.community.detail.failed", err, "community_id", id.Hex())
		return nil, err
	}
	community, err := s.hmd.GetDecentralizedCommunity(ctx, id)
	if err != nil {
		logPublishResult(ctx, "publish.community.detail.success", "publish.community.detail.failed", err, "community_id", id.Hex())
		return nil, err
	}
	if community != nil {
		if err := requireDecentralizedCommunityAccess(ctx, scope, community.ID, "get decentralized community"); err != nil {
			logPublishWarn(ctx, "publish.community.detail.denied", "community_id", id.Hex(), "error", err)
			return nil, err
		}
	}
	logPublishInfo(ctx, "publish.community.detail.success", "community_id", id.Hex(), "found", community != nil)
	return community, nil
}

func (s *decentralizedCommunityService) ListDecentralizedCommunities(ctx context.Context, input ListDecentralizedCommunitiesInput) ([]hmdmodel.HmdDecentralized, error) {
	logPublishInfo(ctx, "publish.community.list.start", "city", input.City, "district", input.District)
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.community.list.success", "publish.community.list.failed", err, "city", input.City, "district", input.District)
		return nil, err
	}
	communities, err := s.hmd.ListDecentralizedCommunities(ctx, input)
	if err != nil {
		logPublishResult(ctx, "publish.community.list.success", "publish.community.list.failed", err, "city", input.City, "district", input.District)
		return nil, err
	}
	filtered, err := filterDecentralizedCommunitiesByScope(ctx, scope, communities)
	if err != nil {
		logPublishResult(ctx, "publish.community.list.success", "publish.community.list.failed", err, "input_count", len(communities))
		return nil, err
	}
	logPublishInfo(ctx, "publish.community.list.success", "input_count", len(communities), "result_count", len(filtered))
	return filtered, nil
}

func (s *decentralizedCommunityService) UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*hmdmodel.HmdDecentralized, error) {
	logPublishInfo(ctx, "publish.community.update.start", "community_id", input.ID.Hex())
	scope, err := newPublishScope(ctx, s.access)
	if err != nil {
		logPublishResult(ctx, "publish.community.update.success", "publish.community.update.failed", err, "community_id", input.ID.Hex())
		return nil, err
	}
	if err := requireDecentralizedCommunityAccess(ctx, scope, input.ID, "update decentralized community"); err != nil {
		logPublishWarn(ctx, "publish.community.update.denied", "community_id", input.ID.Hex(), "error", err)
		return nil, err
	}
	result, err := s.hmd.UpdateDecentralizedCommunity(ctx, input)
	entity, err := resolveHmdMutation(ctx, s.publisher, result, err)
	logPublishResult(ctx, "publish.community.update.success", "publish.community.update.failed", err, "community_id", input.ID.Hex())
	return entity, err
}
