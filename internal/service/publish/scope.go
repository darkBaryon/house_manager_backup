package publish

import (
	"context"
	"fmt"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PublishScope struct {
	principal session.Principal
	access    publishAccessService
}

func newPublishScope(ctx context.Context, access publishAccessService) (PublishScope, error) {
	principal, err := requirePublishPrincipal(ctx)
	if err != nil {
		return PublishScope{}, err
	}
	return PublishScope{
		principal: principal,
		access:    access,
	}, nil
}

func (s PublishScope) accessibleProjectIDSet(ctx context.Context) (map[bson.ObjectID]struct{}, error) {
	if s.access == nil {
		return nil, errcode.InternalError.WithError(fmt.Errorf("发房权限服务未初始化"))
	}
	ids, err := s.access.ListAccessibleProjectIDs(ctx, s.principal)
	if err != nil {
		return nil, err
	}
	set := make(map[bson.ObjectID]struct{}, len(ids))
	for _, id := range ids {
		if id.IsZero() {
			continue
		}
		set[id] = struct{}{}
	}
	return set, nil
}

func (s PublishScope) accessibleProjectIDs(ctx context.Context) ([]bson.ObjectID, error) {
	if s.access == nil {
		return nil, errcode.InternalError.WithError(fmt.Errorf("发房权限服务未初始化"))
	}
	return s.access.ListAccessibleProjectIDs(ctx, s.principal)
}

func (s PublishScope) accessibleCommunityIDSet(ctx context.Context) (map[bson.ObjectID]struct{}, error) {
	if s.access == nil {
		return nil, errcode.InternalError.WithError(fmt.Errorf("发房权限服务未初始化"))
	}
	ids, err := s.access.ListAccessibleCommunityIDs(ctx, s.principal)
	if err != nil {
		return nil, err
	}
	set := make(map[bson.ObjectID]struct{}, len(ids))
	for _, id := range ids {
		if id.IsZero() {
			continue
		}
		set[id] = struct{}{}
	}
	return set, nil
}

func (s PublishScope) accessibleCommunityIDs(ctx context.Context) ([]bson.ObjectID, error) {
	if s.access == nil {
		return nil, errcode.InternalError.WithError(fmt.Errorf("发房权限服务未初始化"))
	}
	return s.access.ListAccessibleCommunityIDs(ctx, s.principal)
}

func scopeNotFound(action string) error {
	return errcode.NotFound.WithError(fmt.Errorf("资源不存在或当前账号无权访问"))
}

func isNotFoundError(err error) bool {
	code := errcode.FromError(err)
	return code != nil && code.Code == errcode.NotFound.Code
}
