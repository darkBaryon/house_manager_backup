package publish

import (
	"context"
	"fmt"
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func requirePublishPrincipal(ctx context.Context) (session.Principal, error) {
	principal, ok := session.PrincipalFromContext(ctx)
	if !ok {
		return session.Principal{}, errcode.Unauthorized.WithError(fmt.Errorf("缺少登录身份信息"))
	}
	if principal.Terminal != session.TerminalPublish {
		return session.Principal{}, errcode.Unauthorized.WithError(fmt.Errorf("当前登录终端不支持该操作"))
	}
	if principal.PrincipalType != session.PrincipalTypeLandlord {
		return session.Principal{}, errcode.Unauthorized.WithError(fmt.Errorf("当前登录身份不是房东"))
	}
	return principal, nil
}

func registerRootScope(ctx context.Context, access publishRootScopeRegistrar, rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, principal session.Principal) error {
	if access == nil {
		return errcode.InternalError.WithError(fmt.Errorf("发房权限服务未初始化"))
	}
	_, err := access.UpsertRootScopeForPrincipal(ctx, rootType, rootID, principal)
	return err
}
