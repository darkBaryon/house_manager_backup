package mongo

import (
	"context"
	"fmt"
)

func (c *Client) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	session, err := c.raw.StartSession()
	if err != nil {
		return fmt.Errorf("start mongo session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		return nil, fn(txCtx)
	})
	if err != nil {
		return fmt.Errorf("mongo transaction: %w", err)
	}
	return nil
}
