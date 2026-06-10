package listingprojection

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"house-manager/internal/domain/hmd"
	commonmodel "house-manager/internal/model/common"
	hmdmodel "house-manager/internal/model/hmd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// recordingFuncs 每个字段被调用时按 scope 名登记，用于断言路由正确性。
func recordingFuncs(calls *[]string) refreshFuncs {
	mk := func(name hmd.HmdProjectionScope) func(context.Context, bson.ObjectID) error {
		return func(context.Context, bson.ObjectID) error {
			*calls = append(*calls, string(name))
			return nil
		}
	}
	return refreshFuncs{
		CentralizedProject:     mk(hmd.HmdScopeCentralizedProject),
		Building:               mk(hmd.HmdScopeBuilding),
		RoomTypeCentralized:    mk(hmd.HmdScopeRoomTypeCentralized),
		CentralizedRoom:        mk(hmd.HmdScopeCentralizedRoom),
		DecentralizedCommunity: mk(hmd.HmdScopeDecentralizedCommunity),
		DecentralizedRoom:      mk(hmd.HmdScopeDecentralizedRoom),
	}
}

// TestDispatchRefreshCoversAllProjectionScopes 遍历唯一枚举源
// hmd.AllProjectionScopes，断言每个 scope 都被路由到对应实体函数且不走 default
// （方案 v2 阻塞 1 要求的必做穷举测试）。
func TestDispatchRefreshCoversAllProjectionScopes(t *testing.T) {
	if len(hmd.AllProjectionScopes) == 0 {
		t.Fatal("hmd.AllProjectionScopes is empty")
	}
	for _, scope := range hmd.AllProjectionScopes {
		t.Run(string(scope), func(t *testing.T) {
			var calls []string
			err := dispatchRefresh(context.Background(), "listingprojection.miniapp.refresh.unhandled_scope",
				hmd.HmdChange{Scope: scope, EntityID: bson.NewObjectID()}, recordingFuncs(&calls))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(calls) != 1 || calls[0] != string(scope) {
				t.Fatalf("scope %q routed to %v, want exactly [%q] (default reached or wrong binding)", scope, calls, scope)
			}
		})
	}
}

// TestDispatchRefreshUnknownScope 断言未知 scope 与旧 switch 语义保真：
// 不调用任何实体函数、返回 nil，并记 unhandled_scope Warn。
func TestDispatchRefreshUnknownScope(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	var calls []string
	err := dispatchRefresh(context.Background(), "listingprojection.miniapp.refresh.unhandled_scope",
		hmd.HmdChange{Scope: "bogus_scope", EntityID: bson.NewObjectID()}, recordingFuncs(&calls))
	if err != nil {
		t.Fatalf("unknown scope should return nil (semantic parity with old switch), got %v", err)
	}
	if len(calls) != 0 {
		t.Fatalf("unknown scope should not invoke any refresh func, got %v", calls)
	}
	if !strings.Contains(buf.String(), "listingprojection.miniapp.refresh.unhandled_scope") {
		t.Fatalf("expected unhandled_scope warn, got log: %s", buf.String())
	}
}

// TestMiniappProjectorRefreshBinding 用真实 projector + fake repo 做一次端到端
// 绑定冒烟：Refresh(centralized_room) 应触发 HPD upsert（绑定声明未接错）。
func TestMiniappProjectorRefreshBinding(t *testing.T) {
	roomID := bson.NewObjectID()
	projectID := bson.NewObjectID()
	buildingID := bson.NewObjectID()

	hpdListings := &fanoutHpdListingRepo{}
	miniappListings := &fanoutMiniappListingRepo{}
	projector := NewMiniappProjector(
		hpdListings,
		miniappListings,
		&fanoutCentralizedRepo{project: &hmdmodel.HmdCentralized{
			CommonFields: commonmodel.CommonFields{ID: projectID},
			ProjectName:  "测试项目",
			City:         "深圳",
		}},
		&fanoutBuildingRepo{building: &hmdmodel.HmdBuilding{
			CommonFields: commonmodel.CommonFields{ID: buildingID},
			ProjectID:    projectID,
			BuildingName: "A座",
		}},
		nil,
		&fanoutRoomTypeRepo{},
		&fanoutCentralizedRoomRepo{rooms: map[bson.ObjectID]*hmdmodel.HmdRoomCentralized{
			roomID: {
				CommonFields: commonmodel.CommonFields{ID: roomID},
				ProjectID:    projectID,
				BuildingID:   buildingID,
				RoomNo:       "801",
				RentMode:     hmdmodel.RentModeWhole,
				Rent:         5200,
				RoomStatus:   hmdmodel.RoomStatusAvailable,
			},
		}},
		nil,
	)

	err := projector.Refresh(context.Background(),
		hmd.HmdChange{Scope: hmd.HmdScopeCentralizedRoom, EntityID: roomID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hpdListings.upsertedSources) != 1 || hpdListings.upsertedSources[0] != roomID {
		t.Fatalf("expected one listing upsert for room, got %#v", hpdListings.upsertedSources)
	}
}
