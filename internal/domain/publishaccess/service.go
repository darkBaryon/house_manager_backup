package publishaccess

import (
	"context"
	hpdmodel "house-manager/internal/model/hpd"
	repohpd "house-manager/internal/repository/hpd"
	"house-manager/pkg/session"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Service owns publish-side owner root scope helpers.
type Service struct {
	rootScopeRepo hpdRootScopeRepository
}

func NewService(
	hpdRootScopeRepo *repohpd.RootScopeRepository,
) *Service {
	return &Service{
		rootScopeRepo: hpdRootScopeRepo,
	}
}

func (s *Service) UpsertRootScopeForPrincipal(ctx context.Context, rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, principal session.Principal) (*hpdmodel.HpdRootScopeRelation, error) {
	if s == nil || s.rootScopeRepo == nil {
		return nil, databasef("upsert hpd root scope relation: root scope repo is nil")
	}
	relation, err := rootScopeRelationForPrincipal(rootType, rootID, principal)
	if err != nil {
		return nil, err
	}
	result, err := s.rootScopeRepo.UpsertActiveByRoot(ctx, relation)
	if err != nil {
		return nil, databasef("upsert hpd root scope relation: %w", err)
	}
	return result, nil
}

func (s *Service) ListAccessibleProjectIDs(ctx context.Context, principal session.Principal) ([]bson.ObjectID, error) {
	return s.listRootIDsForPrincipal(ctx, hpdmodel.HpdRootScopeTypeCentralizedProject, principal, "list accessible centralized project IDs")
}

func (s *Service) ListAccessibleCommunityIDs(ctx context.Context, principal session.Principal) ([]bson.ObjectID, error) {
	return s.listRootIDsForPrincipal(ctx, hpdmodel.HpdRootScopeTypeDecentralizedCommunity, principal, "list accessible decentralized community IDs")
}

func (s *Service) CanAccessProjectForPrincipal(ctx context.Context, projectID bson.ObjectID, principal session.Principal) (bool, error) {
	return s.canAccessRootForPrincipal(ctx, hpdmodel.HpdRootScopeTypeCentralizedProject, projectID, principal, "can access centralized project")
}

func (s *Service) CanAccessCommunityForPrincipal(ctx context.Context, communityID bson.ObjectID, principal session.Principal) (bool, error) {
	return s.canAccessRootForPrincipal(ctx, hpdmodel.HpdRootScopeTypeDecentralizedCommunity, communityID, principal, "can access decentralized community")
}

func (s *Service) listRootIDsForPrincipal(ctx context.Context, rootType hpdmodel.HpdRootScopeType, principal session.Principal, action string) ([]bson.ObjectID, error) {
	if s == nil || s.rootScopeRepo == nil {
		return nil, databasef("%s: root scope repo is nil", action)
	}
	ownerPhone, err := ownerPhoneForPrincipal(principal)
	if err != nil {
		return nil, err
	}
	ids, err := s.rootScopeRepo.ListActiveRootIDsByOwnerPhone(ctx, rootType, ownerPhone)
	if err != nil {
		return nil, databasef("%s: %w", action, err)
	}
	return ids, nil
}

func (s *Service) canAccessRootForPrincipal(ctx context.Context, rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, principal session.Principal, action string) (bool, error) {
	if s == nil || s.rootScopeRepo == nil {
		return false, databasef("%s: root scope repo is nil", action)
	}
	ownerPhone, err := ownerPhoneForPrincipal(principal)
	if err != nil {
		return false, err
	}
	allowed, err := s.rootScopeRepo.CanAccessRoot(ctx, rootType, rootID, ownerPhone)
	if err != nil {
		return false, databasef("%s: %w", action, err)
	}
	return allowed, nil
}

func rootScopeRelationForPrincipal(rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, principal session.Principal) (*hpdmodel.HpdRootScopeRelation, error) {
	if !rootType.Valid() {
		return nil, invalidParamf("root_type is invalid")
	}
	if rootID.IsZero() {
		return nil, invalidParamf("root_id is required")
	}
	ownerPhone, err := ownerPhoneForPrincipal(principal)
	if err != nil {
		return nil, err
	}
	return &hpdmodel.HpdRootScopeRelation{
		RootType:       rootType,
		RootID:         rootID,
		OwnerPhone:     ownerPhone,
		RelationStatus: hpdmodel.HpdRelationStatusActive,
	}, nil
}

func ownerPhoneForPrincipal(principal session.Principal) (string, error) {
	if err := principal.Validate(); err != nil {
		return "", invalidParamf("principal is invalid: %w", err)
	}
	if principal.Terminal != session.TerminalPublish {
		return "", invalidParamf("principal terminal must be publish")
	}
	if principal.PrincipalType != session.PrincipalTypeUser {
		return "", invalidParamf("principal_type must be user")
	}
	ownerPhone := strings.TrimSpace(principal.Phone)
	if ownerPhone == "" {
		return "", invalidParamf("user principal phone is required")
	}
	return ownerPhone, nil
}
