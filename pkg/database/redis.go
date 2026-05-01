package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addrs        []string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	ConnTimeout  int
	ReadTimeout  int
	WriteTimeout int
	MaxRetries   int
	ClusterMode  bool
}

func NewRedisClient(ctx context.Context, cfg RedisConfig) (RedisClient, error) {
	if cfg.ClusterMode {
		return newClusterClient(ctx, cfg)
	}
	return newStandaloneClient(ctx, cfg)
}

type RedisClient interface {
	redis.Cmdable
	Ping(ctx context.Context) *redis.StatusCmd
	Close() error
}

func newStandaloneClient(ctx context.Context, cfg RedisConfig) (RedisClient, error) {
	addr := "localhost:6379"
	if len(cfg.Addrs) > 0 {
		addr = cfg.Addrs[0]
	}

	opts := &redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
	}

	if cfg.ConnTimeout > 0 {
		opts.DialTimeout = time.Duration(cfg.ConnTimeout) * time.Second
	}
	if cfg.ReadTimeout > 0 {
		opts.ReadTimeout = time.Duration(cfg.ReadTimeout) * time.Second
	}
	if cfg.WriteTimeout > 0 {
		opts.WriteTimeout = time.Duration(cfg.WriteTimeout) * time.Second
	}

	rdb := redis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	slog.Info("redis connected", "addr", addr)
	return rdb, nil
}

func newClusterClient(ctx context.Context, cfg RedisConfig) (RedisClient, error) {
	opts := &redis.ClusterOptions{
		Addrs:        cfg.Addrs,
		Password:     cfg.Password,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
	}

	if cfg.ConnTimeout > 0 {
		opts.DialTimeout = time.Duration(cfg.ConnTimeout) * time.Second
	}
	if cfg.ReadTimeout > 0 {
		opts.ReadTimeout = time.Duration(cfg.ReadTimeout) * time.Second
	}
	if cfg.WriteTimeout > 0 {
		opts.WriteTimeout = time.Duration(cfg.WriteTimeout) * time.Second
	}

	rdb := redis.NewClusterClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis cluster: %w", err)
	}

	slog.Info("redis cluster connected", "addrs", cfg.Addrs)
	return rdb, nil
}
