package publishaccess

import (
	"context"
	hpdmodel "house-manager/internal/model/hpd"
	"house-manager/pkg/session"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestServiceUpsertsEntrustForStaffPrincipal(t *testing.T) {
	listingID := bson.NewObjectID()
	staffID := bson.NewObjectID()
	repo := &fakeHpdEntrustRelationRepo{}
	service := &Service{entrustRepo: repo}

	relation, err := service.UpsertEntrustForPrincipal(context.Background(), listingID, session.Principal{
		PrincipalType: session.PrincipalTypeStaff,
		PrincipalID:   staffID.Hex(),
		Terminal:      session.TerminalPublish,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if relation == nil || relation.ListingID != listingID {
		t.Fatalf("unexpected relation: %#v", relation)
	}
	if repo.upserted.MaintainerStaffID != staffID || repo.upserted.ServiceStaffID != staffID {
		t.Fatalf("expected staff relation, got %#v", repo.upserted)
	}
	if repo.upserted.OwnerPhone != "" {
		t.Fatalf("staff relation should not set owner phone")
	}
}

func TestServiceUpsertsEntrustForUserPrincipal(t *testing.T) {
	listingID := bson.NewObjectID()
	repo := &fakeHpdEntrustRelationRepo{}
	service := &Service{entrustRepo: repo}

	_, err := service.UpsertEntrustForPrincipal(context.Background(), listingID, session.Principal{
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
	if !repo.upserted.MaintainerStaffID.IsZero() || !repo.upserted.ServiceStaffID.IsZero() {
		t.Fatalf("user relation should not set staff ids: %#v", repo.upserted)
	}
}

type fakeHpdEntrustRelationRepo struct {
	upserted *hpdmodel.HpdEntrustRelation
}

func (f *fakeHpdEntrustRelationRepo) UpsertActiveByListingID(ctx context.Context, entity *hpdmodel.HpdEntrustRelation) (*hpdmodel.HpdEntrustRelation, error) {
	f.upserted = entity
	return entity, nil
}

func (f *fakeHpdEntrustRelationRepo) FindActiveByListingID(ctx context.Context, listingID bson.ObjectID) (*hpdmodel.HpdEntrustRelation, error) {
	return nil, nil
}

func (f *fakeHpdEntrustRelationRepo) ListActiveListingIDsByStaff(ctx context.Context, staffID bson.ObjectID) ([]bson.ObjectID, error) {
	return nil, nil
}

func (f *fakeHpdEntrustRelationRepo) ListActiveListingIDsByOwnerPhone(ctx context.Context, ownerPhone string) ([]bson.ObjectID, error) {
	return nil, nil
}

func (f *fakeHpdEntrustRelationRepo) CanAccessListing(ctx context.Context, listingID, staffID bson.ObjectID, ownerPhone string) (bool, error) {
	return false, nil
}
