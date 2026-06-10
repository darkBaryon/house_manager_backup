package publishaccess

import (
	"context"
	hpdmodel "house-manager/internal/model/hpd"
	repohpd "house-manager/internal/repository/hpd"
	"house-manager/pkg/applog"
	"house-manager/pkg/session"
	"log/slog"
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
	slog.InfoContext(ctx, "publishaccess.root_scope.upsert.start", "root_type", rootType, "root_id", rootID.Hex())
	if s == nil || s.rootScopeRepo == nil {
		err := databasef("upsert hpd root scope relation: root scope repo is nil")
		applog.Result(ctx, "publishaccess.root_scope.upsert.success", "publishaccess.root_scope.upsert.failed", err, "root_type", rootType, "root_id", rootID.Hex())
		return nil, err
	}
	relation, err := rootScopeRelationForPrincipal(rootType, rootID, principal)
	if err != nil {
		applog.Result(ctx, "publishaccess.root_scope.upsert.success", "publishaccess.root_scope.upsert.failed", err, "root_type", rootType, "root_id", rootID.Hex())
		return nil, err
	}
	result, err := s.rootScopeRepo.UpsertActiveByRoot(ctx, relation)
	if err != nil {
		err = databasef("upsert hpd root scope relation: %w", err)
		applog.Result(ctx, "publishaccess.root_scope.upsert.success", "publishaccess.root_scope.upsert.failed", err, "root_type", rootType, "root_id", rootID.Hex())
		return nil, err
	}
	slog.InfoContext(ctx, "publishaccess.root_scope.upsert.success", "root_type", rootType, "root_id", rootID.Hex(), "owner_phone", applog.Phone(relation.OwnerPhone))
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
	slog.InfoContext(ctx, "publishaccess.root_scope.list.start", "root_type", rootType, "action", action)
	if s == nil || s.rootScopeRepo == nil {
		err := databasef("%s: root scope repo is nil", action)
		applog.Result(ctx, "publishaccess.root_scope.list.success", "publishaccess.root_scope.list.failed", err, "root_type", rootType, "action", action)
		return nil, err
	}
	ownerLandlordID, ownerPhone, err := ownerForPrincipal(principal)
	if err != nil {
		applog.Result(ctx, "publishaccess.root_scope.list.success", "publishaccess.root_scope.list.failed", err, "root_type", rootType, "action", action)
		return nil, err
	}
	ids, err := s.rootScopeRepo.ListActiveRootIDsByOwnerLandlordID(ctx, rootType, ownerLandlordID)
	if err != nil {
		err = databasef("%s: %w", action, err)
		applog.Result(ctx, "publishaccess.root_scope.list.success", "publishaccess.root_scope.list.failed", err, "root_type", rootType, "owner_landlord_id", ownerLandlordID.Hex(), "owner_phone", applog.Phone(ownerPhone))
		return nil, err
	}
	slog.InfoContext(ctx, "publishaccess.root_scope.list.success", "root_type", rootType, "owner_landlord_id", ownerLandlordID.Hex(), "owner_phone", applog.Phone(ownerPhone), "result_count", len(ids))
	return ids, nil
}

func (s *Service) canAccessRootForPrincipal(ctx context.Context, rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, principal session.Principal, action string) (bool, error) {
	slog.InfoContext(ctx, "publishaccess.root_scope.can_access.start", "root_type", rootType, "root_id", rootID.Hex(), "action", action)
	if s == nil || s.rootScopeRepo == nil {
		err := databasef("%s: root scope repo is nil", action)
		applog.Result(ctx, "publishaccess.root_scope.can_access.success", "publishaccess.root_scope.can_access.failed", err, "root_type", rootType, "root_id", rootID.Hex())
		return false, err
	}
	ownerLandlordID, ownerPhone, err := ownerForPrincipal(principal)
	if err != nil {
		applog.Result(ctx, "publishaccess.root_scope.can_access.success", "publishaccess.root_scope.can_access.failed", err, "root_type", rootType, "root_id", rootID.Hex())
		return false, err
	}
	allowed, err := s.rootScopeRepo.CanAccessRoot(ctx, rootType, rootID, ownerLandlordID)
	if err != nil {
		err = databasef("%s: %w", action, err)
		applog.Result(ctx, "publishaccess.root_scope.can_access.success", "publishaccess.root_scope.can_access.failed", err, "root_type", rootType, "root_id", rootID.Hex(), "owner_landlord_id", ownerLandlordID.Hex(), "owner_phone", applog.Phone(ownerPhone))
		return false, err
	}
	slog.InfoContext(ctx, "publishaccess.root_scope.can_access.success", "root_type", rootType, "root_id", rootID.Hex(), "owner_landlord_id", ownerLandlordID.Hex(), "owner_phone", applog.Phone(ownerPhone), "allowed", allowed)
	return allowed, nil
}

func rootScopeRelationForPrincipal(rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, principal session.Principal) (*hpdmodel.HpdRootScopeRelation, error) {
	if !rootType.Valid() {
		return nil, invalidParamf("归属根类型不合法")
	}
	if rootID.IsZero() {
		return nil, invalidParamf("归属根 ID 不能为空")
	}
	ownerLandlordID, ownerPhone, err := ownerForPrincipal(principal)
	if err != nil {
		return nil, err
	}
	return &hpdmodel.HpdRootScopeRelation{
		RootType:        rootType,
		RootID:          rootID,
		OwnerLandlordID: ownerLandlordID,
		OwnerPhone:      ownerPhone,
		RelationStatus:  hpdmodel.HpdRelationStatusActive,
	}, nil
}

func ownerForPrincipal(principal session.Principal) (bson.ObjectID, string, error) {
	if err := principal.Validate(); err != nil {
		return bson.NilObjectID, "", invalidParamf("登录身份无效: %w", err)
	}
	if principal.Terminal != session.TerminalPublish {
		return bson.NilObjectID, "", invalidParamf("当前登录终端不是发房端")
	}
	if principal.PrincipalType != session.PrincipalTypeLandlord {
		return bson.NilObjectID, "", invalidParamf("当前登录身份不是房东")
	}
	ownerLandlordID, err := bson.ObjectIDFromHex(principal.PrincipalID)
	if err != nil || ownerLandlordID.IsZero() {
		return bson.NilObjectID, "", invalidParamf("房东身份 ID 无效")
	}
	ownerPhone := strings.TrimSpace(principal.Phone)
	if ownerPhone == "" {
		return bson.NilObjectID, "", invalidParamf("房东手机号不能为空")
	}
	return ownerLandlordID, ownerPhone, nil
}
