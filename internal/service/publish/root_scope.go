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
		return session.Principal{}, errcode.Unauthorized.WithError(fmt.Errorf("publish principal is required"))
	}
	if principal.Terminal != session.TerminalPublish {
		return session.Principal{}, errcode.Unauthorized.WithError(fmt.Errorf("publish principal is required"))
	}
	return principal, nil
}

func registerRootScope(ctx context.Context, access publishRootScopeRegistrar, rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, principal session.Principal) error {
	if access == nil {
		return errcode.InternalError.WithError(fmt.Errorf("publish access service is required"))
	}
	_, err := access.UpsertRootScopeForPrincipal(ctx, rootType, rootID, principal)
	return err
}
