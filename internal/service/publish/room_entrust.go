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

func registerRoomEntrust(ctx context.Context, entrust publishEntrustRegistrar, sourceType hpdmodel.HpdSourceType, sourceID bson.ObjectID, principal session.Principal) error {
	if entrust == nil {
		return errcode.InternalError.WithError(fmt.Errorf("publish access service is required"))
	}
	listing, err := entrust.FindListingBySource(ctx, sourceType, sourceID)
	if err != nil {
		return err
	}
	if listing == nil {
		return errcode.DatabaseError.WithError(fmt.Errorf("hpd listing not found for %s %s", sourceType, sourceID.Hex()))
	}
	_, err = entrust.UpsertEntrustForPrincipal(ctx, listing.ID, principal)
	return err
}
