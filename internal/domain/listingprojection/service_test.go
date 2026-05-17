package listingprojection

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"house-manager/internal/domain/hmd"
	commonmodel "house-manager/internal/model/common"
	hpdmodel "house-manager/internal/model/hpd"
)

func TestServiceApplyDispatchesAllMiniappScopes(t *testing.T) {
	ids := []bson.ObjectID{
		bson.NewObjectID(),
		bson.NewObjectID(),
		bson.NewObjectID(),
		bson.NewObjectID(),
		bson.NewObjectID(),
		bson.NewObjectID(),
	}
	projector := &fakeMiniappProjector{}
	publisher := &fakePublisherProjector{}
	admin := &fakeAdminProjector{}
	service := &Service{miniappProjector: projector, publisherProjector: publisher, adminProjector: admin}

	err := service.Apply(context.Background(), []hmd.HmdChange{
		{Scope: hmd.HmdScopeCentralizedProject, EntityID: ids[0]},
		{Scope: hmd.HmdScopeBuilding, EntityID: ids[1]},
		{Scope: hmd.HmdScopeRoomTypeCentralized, EntityID: ids[2]},
		{Scope: hmd.HmdScopeCentralizedRoom, EntityID: ids[3]},
		{Scope: hmd.HmdScopeDecentralizedCommunity, EntityID: ids[4]},
		{Scope: hmd.HmdScopeDecentralizedRoom, EntityID: ids[5]},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{
		"centralized_project:" + ids[0].Hex(),
		"building:" + ids[1].Hex(),
		"room_type_centralized:" + ids[2].Hex(),
		"centralized_room:" + ids[3].Hex(),
		"decentralized_community:" + ids[4].Hex(),
		"decentralized_room:" + ids[5].Hex(),
	}
	if len(projector.calls) != len(want) {
		t.Fatalf("expected %d calls, got %#v", len(want), projector.calls)
	}
	for i := range want {
		if projector.calls[i] != want[i] {
			t.Fatalf("call %d: expected %s, got %s", i, want[i], projector.calls[i])
		}
		if publisher.calls[i] != want[i] {
			t.Fatalf("publisher call %d: expected %s, got %s", i, want[i], publisher.calls[i])
		}
		if admin.calls[i] != want[i] {
			t.Fatalf("admin call %d: expected %s, got %s", i, want[i], admin.calls[i])
		}
	}
}

func TestServiceUpdateListingStatusRefreshesMiniappProjection(t *testing.T) {
	listingID := bson.NewObjectID()
	listing := &hpdmodel.HpdListing{
		CommonFields:  commonmodel.CommonFields{ID: listingID},
		SourceType:    hpdmodel.HpdSourceTypeCentralizedRoom,
		SourceID:      bson.NewObjectID(),
		AssetMode:     hpdmodel.HpdAssetModeCentralized,
		ListingStatus: hpdmodel.HpdListingStatusPublished,
	}
	repo := &fakeHpdListingRepo{listing: listing}
	projector := &fakeMiniappProjector{}
	publisher := &fakePublisherProjector{}
	admin := &fakeAdminProjector{}
	service := &Service{listingRepo: repo, miniappProjector: projector, publisherProjector: publisher, adminProjector: admin}

	if err := service.UpdateListingStatus(context.Background(), listingID, hpdmodel.HpdListingStatusPublished); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.updatedStatus != hpdmodel.HpdListingStatusPublished {
		t.Fatalf("expected listing status update, got %d", repo.updatedStatus)
	}
	if len(projector.refreshedListings) != 1 || projector.refreshedListings[0] != listing {
		t.Fatalf("expected listing refresh, got %#v", projector.refreshedListings)
	}
	if len(publisher.refreshedListings) != 1 || publisher.refreshedListings[0] != listing {
		t.Fatalf("expected publisher listing refresh, got %#v", publisher.refreshedListings)
	}
	if len(admin.refreshedListings) != 1 || admin.refreshedListings[0] != listing {
		t.Fatalf("expected admin listing refresh, got %#v", admin.refreshedListings)
	}
}

type fakeMiniappProjector struct {
	calls             []string
	refreshedListings []*hpdmodel.HpdListing
}

type fakePublisherProjector struct {
	calls             []string
	refreshedListings []*hpdmodel.HpdListing
}

type fakeAdminProjector struct {
	calls             []string
	refreshedListings []*hpdmodel.HpdListing
}

func (f *fakeAdminProjector) RefreshByListing(ctx context.Context, listing *hpdmodel.HpdListing) error {
	f.refreshedListings = append(f.refreshedListings, listing)
	return nil
}

func (f *fakeAdminProjector) RefreshCentralizedRoom(ctx context.Context, roomID bson.ObjectID) error {
	f.calls = append(f.calls, "centralized_room:"+roomID.Hex())
	return nil
}

func (f *fakeAdminProjector) RefreshDecentralizedRoom(ctx context.Context, roomID bson.ObjectID) error {
	f.calls = append(f.calls, "decentralized_room:"+roomID.Hex())
	return nil
}

func (f *fakeAdminProjector) RefreshCentralizedProject(ctx context.Context, projectID bson.ObjectID) error {
	f.calls = append(f.calls, "centralized_project:"+projectID.Hex())
	return nil
}

func (f *fakeAdminProjector) RefreshBuilding(ctx context.Context, buildingID bson.ObjectID) error {
	f.calls = append(f.calls, "building:"+buildingID.Hex())
	return nil
}

func (f *fakeAdminProjector) RefreshRoomTypeCentralized(ctx context.Context, roomTypeID bson.ObjectID) error {
	f.calls = append(f.calls, "room_type_centralized:"+roomTypeID.Hex())
	return nil
}

func (f *fakeAdminProjector) RefreshDecentralizedCommunity(ctx context.Context, decentralizedID bson.ObjectID) error {
	f.calls = append(f.calls, "decentralized_community:"+decentralizedID.Hex())
	return nil
}

func (f *fakePublisherProjector) RefreshByListing(ctx context.Context, listing *hpdmodel.HpdListing) error {
	f.refreshedListings = append(f.refreshedListings, listing)
	return nil
}

func (f *fakePublisherProjector) RefreshCentralizedRoom(ctx context.Context, roomID bson.ObjectID) error {
	f.calls = append(f.calls, "centralized_room:"+roomID.Hex())
	return nil
}

func (f *fakePublisherProjector) RefreshDecentralizedRoom(ctx context.Context, roomID bson.ObjectID) error {
	f.calls = append(f.calls, "decentralized_room:"+roomID.Hex())
	return nil
}

func (f *fakePublisherProjector) RefreshCentralizedProject(ctx context.Context, projectID bson.ObjectID) error {
	f.calls = append(f.calls, "centralized_project:"+projectID.Hex())
	return nil
}

func (f *fakePublisherProjector) RefreshBuilding(ctx context.Context, buildingID bson.ObjectID) error {
	f.calls = append(f.calls, "building:"+buildingID.Hex())
	return nil
}

func (f *fakePublisherProjector) RefreshRoomTypeCentralized(ctx context.Context, roomTypeID bson.ObjectID) error {
	f.calls = append(f.calls, "room_type_centralized:"+roomTypeID.Hex())
	return nil
}

func (f *fakePublisherProjector) RefreshDecentralizedCommunity(ctx context.Context, decentralizedID bson.ObjectID) error {
	f.calls = append(f.calls, "decentralized_community:"+decentralizedID.Hex())
	return nil
}

func (f *fakeMiniappProjector) RefreshByListing(ctx context.Context, listing *hpdmodel.HpdListing) error {
	f.refreshedListings = append(f.refreshedListings, listing)
	return nil
}

func (f *fakeMiniappProjector) RefreshCentralizedRoom(ctx context.Context, roomID bson.ObjectID) error {
	f.calls = append(f.calls, "centralized_room:"+roomID.Hex())
	return nil
}

func (f *fakeMiniappProjector) RefreshDecentralizedRoom(ctx context.Context, roomID bson.ObjectID) error {
	f.calls = append(f.calls, "decentralized_room:"+roomID.Hex())
	return nil
}

func (f *fakeMiniappProjector) RefreshCentralizedProject(ctx context.Context, projectID bson.ObjectID) error {
	f.calls = append(f.calls, "centralized_project:"+projectID.Hex())
	return nil
}

func (f *fakeMiniappProjector) RefreshBuilding(ctx context.Context, buildingID bson.ObjectID) error {
	f.calls = append(f.calls, "building:"+buildingID.Hex())
	return nil
}

func (f *fakeMiniappProjector) RefreshRoomTypeCentralized(ctx context.Context, roomTypeID bson.ObjectID) error {
	f.calls = append(f.calls, "room_type_centralized:"+roomTypeID.Hex())
	return nil
}

func (f *fakeMiniappProjector) RefreshDecentralizedCommunity(ctx context.Context, decentralizedID bson.ObjectID) error {
	f.calls = append(f.calls, "decentralized_community:"+decentralizedID.Hex())
	return nil
}

type fakeHpdListingRepo struct {
	listing       *hpdmodel.HpdListing
	updatedStatus hpdmodel.HpdListingStatus
}

func (f *fakeHpdListingRepo) FindByID(ctx context.Context, id bson.ObjectID) (*hpdmodel.HpdListing, error) {
	return f.listing, nil
}

func (f *fakeHpdListingRepo) FindBySource(ctx context.Context, sourceType hpdmodel.HpdSourceType, sourceID bson.ObjectID) (*hpdmodel.HpdListing, error) {
	return f.listing, nil
}

func (f *fakeHpdListingRepo) UpsertBySource(ctx context.Context, entity *hpdmodel.HpdListing) (*hpdmodel.HpdListing, error) {
	return f.listing, nil
}

func (f *fakeHpdListingRepo) UpdateLifecycleFields(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	return nil
}

func (f *fakeHpdListingRepo) UpdateStatus(ctx context.Context, id bson.ObjectID, listingStatus hpdmodel.HpdListingStatus) error {
	f.updatedStatus = listingStatus
	return nil
}
