package listingprojection

import (
	"context"
	"errors"
	"strings"
	"testing"

	"house-manager/internal/domain/hmd"
	commonmodel "house-manager/internal/model/common"
	hpdmodel "house-manager/internal/model/hpd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// fakeProjector 实现窄接口 projector；calls 共享一份序列，
// 用于断言 Apply 的路由顺序（装配顺序 × change 顺序）与 fail-fast。
type fakeProjector struct {
	name              string
	log               *[]string
	failOnScope       hmd.HmdProjectionScope
	failErr           error
	refreshedListings []*hpdmodel.HpdListing
}

func (f *fakeProjector) Refresh(ctx context.Context, change hmd.HmdChange) error {
	*f.log = append(*f.log, f.name+":"+string(change.Scope)+":"+change.EntityID.Hex())
	if f.failOnScope != "" && change.Scope == f.failOnScope {
		return f.failErr
	}
	return nil
}

func (f *fakeProjector) RefreshByListing(ctx context.Context, listing *hpdmodel.HpdListing) error {
	f.refreshedListings = append(f.refreshedListings, listing)
	return nil
}

// TestServiceApplyRoutesAllProjectorsInOrder 是原 TestServiceApplyDispatchesAllMiniappScopes
// 的重定向（方案 v2 阻塞 3）：scope→实体方法的断言已由 refresh_dispatch_test 接手，
// Service 层只断言「每个 change 按装配顺序路由到全部 projectors」。
func TestServiceApplyRoutesAllProjectorsInOrder(t *testing.T) {
	var log []string
	m := &fakeProjector{name: "miniapp", log: &log}
	p := &fakeProjector{name: "publisher", log: &log}
	a := &fakeProjector{name: "admin", log: &log}
	service := &Service{projectors: []projector{m, p, a}}

	id1 := bson.NewObjectID()
	id2 := bson.NewObjectID()
	err := service.Apply(context.Background(), []hmd.HmdChange{
		{Scope: hmd.HmdScopeBuilding, EntityID: id1},
		{Scope: hmd.HmdScopeCentralizedRoom, EntityID: id2},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{
		"miniapp:building:" + id1.Hex(),
		"publisher:building:" + id1.Hex(),
		"admin:building:" + id1.Hex(),
		"miniapp:centralized_room:" + id2.Hex(),
		"publisher:centralized_room:" + id2.Hex(),
		"admin:centralized_room:" + id2.Hex(),
	}
	if len(log) != len(want) {
		t.Fatalf("expected %d calls, got %#v", len(want), log)
	}
	for i := range want {
		if log[i] != want[i] {
			t.Fatalf("call %d: expected %s, got %s", i, want[i], log[i])
		}
	}
}

// TestServiceApplyFailFast 断言首错即停（行为保持）且错误包装含 scope。
func TestServiceApplyFailFast(t *testing.T) {
	var log []string
	m := &fakeProjector{name: "miniapp", log: &log}
	p := &fakeProjector{name: "publisher", log: &log, failOnScope: hmd.HmdScopeCentralizedRoom, failErr: errors.New("boom")}
	a := &fakeProjector{name: "admin", log: &log}
	service := &Service{projectors: []projector{m, p, a}}

	err := service.Apply(context.Background(), []hmd.HmdChange{
		{Scope: hmd.HmdScopeCentralizedRoom, EntityID: bson.NewObjectID()},
		{Scope: hmd.HmdScopeBuilding, EntityID: bson.NewObjectID()},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "centralized_room") {
		t.Fatalf("error should contain scope, got: %v", err)
	}
	// miniapp 成功、publisher 失败后立即停止：admin 与第二个 change 不应被调用
	if len(log) != 2 {
		t.Fatalf("expected fail-fast after 2 calls, got %#v", log)
	}
}

func TestServiceUpdateListingStatusRefreshesAllProjections(t *testing.T) {
	listingID := bson.NewObjectID()
	listing := &hpdmodel.HpdListing{
		CommonFields:  commonmodel.CommonFields{ID: listingID},
		SourceType:    hpdmodel.HpdSourceTypeCentralizedRoom,
		SourceID:      bson.NewObjectID(),
		AssetMode:     hpdmodel.HpdAssetModeCentralized,
		ListingStatus: hpdmodel.HpdListingStatusPublished,
	}
	repo := &fakeHpdListingRepo{listing: listing}
	var log []string
	m := &fakeProjector{name: "miniapp", log: &log}
	p := &fakeProjector{name: "publisher", log: &log}
	a := &fakeProjector{name: "admin", log: &log}
	service := &Service{listingRepo: repo, projectors: []projector{m, p, a}}

	if err := service.UpdateListingStatus(context.Background(), listingID, hpdmodel.HpdListingStatusPublished); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.updatedStatus != hpdmodel.HpdListingStatusPublished {
		t.Fatalf("expected listing status update, got %d", repo.updatedStatus)
	}
	for _, f := range []*fakeProjector{m, p, a} {
		if len(f.refreshedListings) != 1 || f.refreshedListings[0] != listing {
			t.Fatalf("expected %s listing refresh, got %#v", f.name, f.refreshedListings)
		}
	}
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
