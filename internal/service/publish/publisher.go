package publish

import (
	"context"

	hmddomain "house-manager/internal/domain/hmd"
)

type mutationPublisher struct {
	hpd hpdApplier
}

func (p mutationPublisher) Apply(ctx context.Context, changes []hmddomain.HmdChange) error {
	if p.hpd == nil {
		return nil
	}
	return p.hpd.Apply(ctx, changes)
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
