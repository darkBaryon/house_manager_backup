package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	appconfig "house-manager/internal/config"

	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	raw goredis.UniversalClient
}

func NewClient(ctx context.Context, cfg appconfig.RedisConfig) (*Client, error) {
	if cfg.ClusterMode {
		return newClusterClient(ctx, cfg)
	}
	return newStandaloneClient(ctx, cfg)
}

func (c *Client) Raw() goredis.UniversalClient {
	return c.raw
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.raw.Get(ctx, key).Result()
	if err == goredis.Nil {
		return "", err
	}
	return val, err
}

func (c *Client) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.raw.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.raw.Del(ctx, keys...).Err()
}

func (c *Client) Ping(ctx context.Context) error {
	if err := c.raw.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	return nil
}

func (c *Client) Close() error {
	if err := c.raw.Close(); err != nil {
		return fmt.Errorf("close redis: %w", err)
	}
	return nil
}

func newStandaloneClient(ctx context.Context, cfg appconfig.RedisConfig) (*Client, error) {
	if len(cfg.Addrs) == 0 || cfg.Addrs[0] == "" {
		return nil, fmt.Errorf("redis addr is required")
	}
	addr := cfg.Addrs[0]

	opts := &goredis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
	}
	applyStandaloneTimeouts(opts, cfg)

	raw := goredis.NewClient(opts)
	if err := raw.Ping(ctx).Err(); err != nil {
		_ = raw.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	slog.Info("redis connected", "addr", addr)
	return &Client{raw: raw}, nil
}

func newClusterClient(ctx context.Context, cfg appconfig.RedisConfig) (*Client, error) {
	if len(cfg.Addrs) == 0 {
		return nil, fmt.Errorf("redis cluster addrs are required")
	}

	opts := &goredis.ClusterOptions{
		Addrs:        cfg.Addrs,
		Password:     cfg.Password,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
	}
	applyClusterTimeouts(opts, cfg)

	raw := goredis.NewClusterClient(opts)
	if err := raw.Ping(ctx).Err(); err != nil {
		_ = raw.Close()
		return nil, fmt.Errorf("ping redis cluster: %w", err)
	}

	slog.Info("redis cluster connected", "addrs", cfg.Addrs)
	return &Client{raw: raw}, nil
}

func applyStandaloneTimeouts(opts *goredis.Options, cfg appconfig.RedisConfig) {
	if cfg.ConnTimeout > 0 {
		opts.DialTimeout = time.Duration(cfg.ConnTimeout) * time.Second
	}
	if cfg.ReadTimeout > 0 {
		opts.ReadTimeout = time.Duration(cfg.ReadTimeout) * time.Second
	}
	if cfg.WriteTimeout > 0 {
		opts.WriteTimeout = time.Duration(cfg.WriteTimeout) * time.Second
	}
}

func applyClusterTimeouts(opts *goredis.ClusterOptions, cfg appconfig.RedisConfig) {
	if cfg.ConnTimeout > 0 {
		opts.DialTimeout = time.Duration(cfg.ConnTimeout) * time.Second
	}
	if cfg.ReadTimeout > 0 {
		opts.ReadTimeout = time.Duration(cfg.ReadTimeout) * time.Second
	}
	if cfg.WriteTimeout > 0 {
		opts.WriteTimeout = time.Duration(cfg.WriteTimeout) * time.Second
	}
}
