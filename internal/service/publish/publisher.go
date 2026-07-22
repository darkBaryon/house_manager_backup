package publish

import (
	"context"
	"fmt"
	"house-manager/pkg/applog"
	"log/slog"

	hmddomain "house-manager/internal/domain/hmd"
	"house-manager/pkg/errcode"
)

type mutationPublisher struct {
	listingProjection listingProjectionApplier
}

func (p mutationPublisher) Apply(ctx context.Context, changes []hmddomain.HmdChange) error {
	if p.listingProjection == nil {
		return errcode.InternalError.WithError(fmt.Errorf("房源投影服务未初始化"))
	}
	slog.InfoContext(ctx, "publish.projection.apply.start", "change_count", len(changes))
	return p.listingProjection.Apply(ctx, changes)
}

func resolveHmdMutation[T any](ctx context.Context, publisher mutationPublisher, result *hmddomain.HmdMutationResult[T], err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	if err := publisher.Apply(ctx, result.Changes); err != nil {
		applog.Result(ctx, "publish.projection.apply.success", "publish.projection.apply.failed", err, "change_count", len(result.Changes))
		return nil, err
	}
	slog.InfoContext(ctx, "publish.projection.apply.success", "change_count", len(result.Changes))
	return result.Entity, nil
}
