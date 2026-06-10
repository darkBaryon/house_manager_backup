package listingprojection

import (
	"context"
	"log/slog"

	"house-manager/internal/domain/hmd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// refreshFuncs 把投影 scope 绑定到某个 projector 的实体刷新方法。
// dispatch 逻辑三端共用（避免三份同构 switch），绑定关系由各 projector
// 的 Refresh 方法以结构体字面量声明，编译器保证签名匹配。
type refreshFuncs struct {
	CentralizedProject     func(context.Context, bson.ObjectID) error
	Building               func(context.Context, bson.ObjectID) error
	RoomTypeCentralized    func(context.Context, bson.ObjectID) error
	CentralizedRoom        func(context.Context, bson.ObjectID) error
	DecentralizedCommunity func(context.Context, bson.ObjectID) error
	DecentralizedRoom      func(context.Context, bson.ObjectID) error
}

// dispatchRefresh 按 change.Scope 路由到对应实体刷新函数。
// 未知 scope 返回 nil：与旧 Apply switch（无 default）的静默跳过语义保真，
// 仅补充 Warn 可观测性（事件 listingprojection.<terminal>.refresh.unhandled_scope）。
func dispatchRefresh(ctx context.Context, terminal string, change hmd.HmdChange, funcs refreshFuncs) error {
	switch change.Scope {
	case hmd.HmdScopeCentralizedProject:
		return funcs.CentralizedProject(ctx, change.EntityID)
	case hmd.HmdScopeBuilding:
		return funcs.Building(ctx, change.EntityID)
	case hmd.HmdScopeRoomTypeCentralized:
		return funcs.RoomTypeCentralized(ctx, change.EntityID)
	case hmd.HmdScopeCentralizedRoom:
		return funcs.CentralizedRoom(ctx, change.EntityID)
	case hmd.HmdScopeDecentralizedCommunity:
		return funcs.DecentralizedCommunity(ctx, change.EntityID)
	case hmd.HmdScopeDecentralizedRoom:
		return funcs.DecentralizedRoom(ctx, change.EntityID)
	default:
		slog.WarnContext(ctx, "listingprojection."+terminal+".refresh.unhandled_scope", "scope", change.Scope, "entity_id", change.EntityID.Hex())
		return nil
	}
}
