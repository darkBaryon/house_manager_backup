package hmd

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
	"strings"
)

func (s *Service) CreateDecentralizedCommunity(ctx context.Context, input CreateDecentralizedCommunityInput) (*HmdMutationResult[hmdmodel.HmdDecentralized], error) {
	entity := &hmdmodel.HmdDecentralized{
		CommunityName: strings.TrimSpace(input.CommunityName),
		City:          strings.TrimSpace(input.City),
		District:      strings.TrimSpace(input.District),
		BizArea:       strings.TrimSpace(input.BizArea),
		AddressText:   strings.TrimSpace(input.AddressText),
		Geo:           toGeoPoint(input.Geo),
		SubwayStation: strings.TrimSpace(input.SubwayStation),
	}
	if err := entity.ValidateForCreate(); err != nil {
		return nil, mutationError("create decentralized community", err)
	}

	existing, err := s.decentralizedRepo.FindByCommunity(ctx, entity.City, entity.District, entity.CommunityName)
	if err != nil {
		return nil, databasef("create decentralized community: find existing community: %w", err)
	}
	if existing != nil {
		return nil, alreadyExistsf("create decentralized community: community already exists in city and district")
	}

	if err := s.decentralizedRepo.Create(ctx, entity); err != nil {
		return nil, mutationError("create decentralized community", err)
	}
	return hmdMutationResult(entity, HmdChange{
		Action:          HmdChangeCreated,
		EntityType:      HmdEntityDecentralizedCommunity,
		EntityID:        entity.ID,
		Scope:           HmdScopeDecentralizedCommunity,
		DecentralizedID: entity.ID,
	}), nil
}

func (s *Service) RollbackDecentralizedCommunityCreate(ctx context.Context, id bson.ObjectID) error {
	if err := s.decentralizedRepo.SoftDeleteByID(ctx, id); err != nil {
		return mutationError("rollback decentralized community create", err)
	}
	return nil
}

func (s *Service) GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdDecentralized, error) {
	return s.requireDecentralized(ctx, id, "get decentralized community")
}

func (s *Service) ListDecentralizedCommunities(ctx context.Context, input ListDecentralizedCommunitiesInput) ([]hmdmodel.HmdDecentralized, error) {
	city := strings.TrimSpace(input.City)
	district := strings.TrimSpace(input.District)
	communities, err := s.decentralizedRepo.List(ctx, city, district)
	if err != nil {
		return nil, databasef("list decentralized communities: %w", err)
	}
	return communities, nil
}

func (s *Service) UpdateDecentralizedCommunity(ctx context.Context, input UpdateDecentralizedCommunityInput) (*HmdMutationResult[hmdmodel.HmdDecentralized], error) {
	community, err := s.requireDecentralized(ctx, input.ID, "update decentralized community")
	if err != nil {
		return nil, err
	}

	city := strings.TrimSpace(input.City)
	district := strings.TrimSpace(input.District)
	communityName := strings.TrimSpace(input.CommunityName)
	existing, err := s.decentralizedRepo.FindByCommunity(ctx, city, district, communityName)
	if err != nil {
		return nil, databasef("update decentralized community: find existing community: %w", err)
	}
	if existing != nil && existing.ID != community.ID {
		return nil, alreadyExistsf("update decentralized community: community already exists in city and district")
	}

	fields := bsonFields(
		"community_name", communityName,
		"city", city,
		"district", district,
		"biz_area", strings.TrimSpace(input.BizArea),
		"address_text", strings.TrimSpace(input.AddressText),
		"geo", toGeoPoint(input.Geo),
		"subway_station", strings.TrimSpace(input.SubwayStation),
	)
	if err := s.decentralizedRepo.UpdateBaseInfo(ctx, community.ID, fields); err != nil {
		return nil, mutationError("update decentralized community", err)
	}
	updated, err := s.requireDecentralized(ctx, community.ID, "update decentralized community")
	if err != nil {
		return nil, err
	}
	return hmdMutationResult(updated, HmdChange{
		Action:          HmdChangeUpdated,
		EntityType:      HmdEntityDecentralizedCommunity,
		EntityID:        updated.ID,
		Scope:           HmdScopeDecentralizedCommunity,
		DecentralizedID: updated.ID,
	}), nil
}
