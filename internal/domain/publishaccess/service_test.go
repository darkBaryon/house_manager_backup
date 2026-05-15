package publishaccess

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/pkg/session"
)

func TestServiceUpsertsRootScopeForUserPrincipal(t *testing.T) {
	rootID := bson.NewObjectID()
	repo := &fakeHpdRootScopeRepo{}
	service := &Service{rootScopeRepo: repo}

	_, err := service.UpsertRootScopeForPrincipal(context.Background(), hpdmodel.HpdRootScopeTypeCentralizedProject, rootID, session.Principal{
		PrincipalType: session.PrincipalTypeUser,
		PrincipalID:   bson.NewObjectID().Hex(),
		Terminal:      session.TerminalPublish,
		Phone:         "13800000000",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.upserted.OwnerPhone != "13800000000" {
		t.Fatalf("expected owner phone relation, got %#v", repo.upserted)
	}
	if repo.upserted.RootType != hpdmodel.HpdRootScopeTypeCentralizedProject || repo.upserted.RootID != rootID {
		t.Fatalf("expected project root scope relation, got %#v", repo.upserted)
	}
}

type fakeHpdRootScopeRepo struct {
	upserted *hpdmodel.HpdRootScopeRelation
}

func (f *fakeHpdRootScopeRepo) UpsertActiveByRoot(ctx context.Context, entity *hpdmodel.HpdRootScopeRelation) (*hpdmodel.HpdRootScopeRelation, error) {
	f.upserted = entity
	return entity, nil
}

func (f *fakeHpdRootScopeRepo) ListActiveRootIDsByOwnerPhone(ctx context.Context, rootType hpdmodel.HpdRootScopeType, ownerPhone string) ([]bson.ObjectID, error) {
	return nil, nil
}

func (f *fakeHpdRootScopeRepo) CanAccessRoot(ctx context.Context, rootType hpdmodel.HpdRootScopeType, rootID bson.ObjectID, ownerPhone string) (bool, error) {
	return false, nil
}
