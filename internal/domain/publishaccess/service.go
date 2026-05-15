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
	logPublishAccessInfo(ctx, "publishaccess.root_scope.upsert.start", "root_type", rootType, "root_id", rootID.Hex())
	if s == nil || s.rootScopeRepo == nil {
		err := databasef("upsert hpd root scope relation: root scope repo is nil")
		logPublishAccessResult(ctx, "publishaccess.root_scope.upsert.success", "publishaccess.root_scope.upsert.failed", err, "root_type", rootType, "root_id", rootID.Hex())
		return nil, err
	}
	relation, err := rootScopeRelationForPrincipal(rootType, rootID, principal)
	if err != nil {
		logPublishAccessResult(ctx, "publishaccess.root_scope.upsert.success", "publishaccess.root_scope.upsert.failed", err, "root_type", rootType, "root_id", rootID.Hex())
		return nil, err
	}
	result, err := s.rootScopeRepo.UpsertActiveByRoot(ctx, relation)
	if err != nil {
		err = databasef("upsert hpd root scope relation: %w", err)
		logPublishAccessResult(ctx, "publishaccess.root_scope.upsert.success", "publishaccess.root_scope.upsert.failed", err, "root_type", rootType, "root_id", rootID.Hex())
		return nil, err
	}
	logPublishAccessInfo(ctx, "publishaccess.root_scope.upsert.success", "root_type", rootType, "root_id", rootID.Hex(), "owner_phone", maskAccessPhone(relation.OwnerPhone))
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
	logPublishAccessInfo(ctx, "publishaccess.root_scope.list.start", "root_type", rootType, "action", action)
	if s == nil || s.rootScopeRepo == nil {
		err := databasef("%s: root scope repo is nil", action)
		logPublishAccessResult(ctx, "publishaccess.root_scope.list.success", "publishaccess.root_scope.list.failed", err, "root_type", rootType, "action", action)
		return nil, err
	}
	ownerPhone, err := ownerPhoneForPrincipal(principal)
	if err != nil {
		logPublishAccessResult(ctx, "publishaccess.root_scope.list.success", "publishaccess.root_scope.list.failed", err, "root_type", rootType, "action", action)
		return nil, err
	}
	ids, err := s.rootScopeRepo.ListActiveRootIDsByOwnerPhone(ctx, rootType, ownerPhone)
	if err != nil {
		err = databasef("%s: %w", action, err)
		logPublishAccessResult(ctx, "publishaccess.root_scope.list.success", "publishaccess.root_scope.list.failed", err, "root_type", rootType, "owner_phone", maskAccessPhone(ownerPhone))
		return nil, err
	}
	logPublishAccessInfo(ctx, "publishaccess.root_scope.list.success", "root_type", rootType, "owner_phone", maskAccessPhone(ownerPhone), "result_count", len(ids))
	return ids, nil
}

func (s *Service) canAccessRootForPrincipal(ctx context.Context, rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, principal session.Principal, action string) (bool, error) {
	logPublishAccessInfo(ctx, "publishaccess.root_scope.can_access.start", "root_type", rootType, "root_id", rootID.Hex(), "action", action)
	if s == nil || s.rootScopeRepo == nil {
		err := databasef("%s: root scope repo is nil", action)
		logPublishAccessResult(ctx, "publishaccess.root_scope.can_access.success", "publishaccess.root_scope.can_access.failed", err, "root_type", rootType, "root_id", rootID.Hex())
		return false, err
	}
	ownerPhone, err := ownerPhoneForPrincipal(principal)
	if err != nil {
		logPublishAccessResult(ctx, "publishaccess.root_scope.can_access.success", "publishaccess.root_scope.can_access.failed", err, "root_type", rootType, "root_id", rootID.Hex())
		return false, err
	}
	allowed, err := s.rootScopeRepo.CanAccessRoot(ctx, rootType, rootID, ownerPhone)
	if err != nil {
		err = databasef("%s: %w", action, err)
		logPublishAccessResult(ctx, "publishaccess.root_scope.can_access.success", "publishaccess.root_scope.can_access.failed", err, "root_type", rootType, "root_id", rootID.Hex(), "owner_phone", maskAccessPhone(ownerPhone))
		return false, err
	}
	logPublishAccessInfo(ctx, "publishaccess.root_scope.can_access.success", "root_type", rootType, "root_id", rootID.Hex(), "owner_phone", maskAccessPhone(ownerPhone), "allowed", allowed)
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
