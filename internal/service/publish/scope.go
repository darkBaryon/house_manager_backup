package publish

import (
	"context"
	"fmt"
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	publishGlobalRole       = "super_admin"
	publishGlobalPermission = "house.manage"
)

type PublishScope struct {
	principal session.Principal
	global    bool
	access    publishAccessService
}

func newPublishScope(ctx context.Context, access publishAccessService) (PublishScope, error) {
	principal, err := requirePublishPrincipal(ctx)
	if err != nil {
		return PublishScope{}, err
	}
	return PublishScope{
		principal: principal,
		global:    hasGlobalPublishAccess(principal),
		access:    access,
	}, nil
}

func (s PublishScope) IsGlobal() bool {
	return s.global
}

func (s PublishScope) requireGlobal(action string) error {
	if s.global {
		return nil
	}
	return scopeNotFound(action)
}

func (s PublishScope) requireCanAccessSource(ctx context.Context, sourceType hpdmodel.HpdSourceType, sourceID bson.ObjectID, action string) error {
	if s.global {
		return nil
	}
	if s.access == nil {
		return errcode.InternalError.WithError(fmt.Errorf("%s: publish access service is required", action))
	}
	allowed, err := s.access.CanAccessSourceForPrincipal(ctx, sourceType, sourceID, s.principal)
	if err != nil {
		return err
	}
	if !allowed {
		return scopeNotFound(action)
	}
	return nil
}

func (s PublishScope) accessibleSourceIDSet(ctx context.Context, sourceType hpdmodel.HpdSourceType) (map[bson.ObjectID]struct{}, error) {
	ids := map[bson.ObjectID]struct{}{}
	if s.global {
		return ids, nil
	}
	if s.access == nil {
		return nil, errcode.InternalError.WithError(fmt.Errorf("publish access service is required"))
	}
	listings, err := s.access.ListAccessibleListings(ctx, s.principal)
	if err != nil {
		return nil, err
	}
	for _, listing := range listings {
		if listing.SourceType != sourceType || listing.SourceID.IsZero() {
			continue
		}
		ids[listing.SourceID] = struct{}{}
	}
	return ids, nil
}

func hasGlobalPublishAccess(principal session.Principal) bool {
	for _, code := range principal.RoleCodes {
		if strings.TrimSpace(code) == publishGlobalRole {
			return true
		}
	}
	for _, code := range principal.PermissionCodes {
		if strings.TrimSpace(code) == publishGlobalPermission {
			return true
		}
	}
	return false
}

func scopeNotFound(action string) error {
	return errcode.NotFound.WithError(fmt.Errorf("%s: resource not found", action))
}

func isNotFoundError(err error) bool {
	code := errcode.FromError(err)
	return code != nil && code.Code == errcode.NotFound.Code
}
