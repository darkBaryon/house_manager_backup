package publish

import (
	"context"
	"fmt"

	hmddomain "house-manager/internal/domain/hmd"
	"house-manager/pkg/errcode"
)

type mutationPublisher struct {
	listingProjection listingProjectionApplier
}

func (p mutationPublisher) Apply(ctx context.Context, changes []hmddomain.HmdChange) error {
	if p.listingProjection == nil {
		return errcode.InternalError.WithError(fmt.Errorf("listing projection service is required"))
	}
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
		return nil, err
	}
	return result.Entity, nil
}
