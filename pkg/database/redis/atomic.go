package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

func (c *Client) Watch(ctx context.Context, fn func(tx *goredis.Tx) error, keys ...string) error {
	if err := c.raw.Watch(ctx, fn, keys...); err != nil {
		return fmt.Errorf("redis watch: %w", err)
	}
	return nil
}

func (c *Client) TxPipelined(ctx context.Context, fn func(pipe goredis.Pipeliner) error) ([]goredis.Cmder, error) {
	cmds, err := c.raw.TxPipelined(ctx, fn)
	if err != nil {
		return nil, fmt.Errorf("redis tx pipeline: %w", err)
	}
	return cmds, nil
}
