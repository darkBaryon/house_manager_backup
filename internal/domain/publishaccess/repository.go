package publishaccess

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	hpdmodel "house-manager/internal/model/hpd"
)

type hpdRootScopeRepository interface {
	UpsertActiveByRoot(ctx context.Context, entity *hpdmodel.HpdRootScopeRelation) (*hpdmodel.HpdRootScopeRelation, error)
	ListActiveRootIDsByOwnerPhone(ctx context.Context, rootType hpdmodel.HpdRootScopeType, ownerPhone string) ([]bson.ObjectID, error)
	CanAccessRoot(ctx context.Context, rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, ownerPhone string) (bool, error)
}
