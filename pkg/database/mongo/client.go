package mongo

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	appconfig "house-manager/internal/config"

	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func uri(cfg appconfig.MongoConfig) string {
	return "mongodb://" + strings.Join(cfg.Addrs, ",")
}

type Client struct {
	raw *drivermongo.Client
	db  *drivermongo.Database
}

func NewClient(ctx context.Context, cfg appconfig.MongoConfig) (*Client, error) {
	opts := options.Client().ApplyURI(uri(cfg))

	if cfg.Username != "" {
		opts.SetAuth(options.Credential{
			AuthSource: cfg.AuthSource,
			Username:   cfg.Username,
			Password:   cfg.Password,
		})
	}
	if cfg.ReplicaSet != "" {
		opts.SetReplicaSet(cfg.ReplicaSet)
	}
	if cfg.PoolSize > 0 {
		opts.SetMaxPoolSize(cfg.PoolSize)
	}
	if cfg.MinPoolSize > 0 {
		opts.SetMinPoolSize(cfg.MinPoolSize)
	}
	if cfg.ConnectTimeout > 0 {
		opts.SetConnectTimeout(time.Duration(cfg.ConnectTimeout) * time.Second)
	}
	if cfg.ServerSelectionTimeout > 0 {
		opts.SetServerSelectionTimeout(time.Duration(cfg.ServerSelectionTimeout) * time.Second)
	}
	if cfg.MaxRetries > 0 {
		opts.SetRetryWrites(true)
		opts.SetRetryReads(true)
	}

	raw, err := drivermongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("connect mongodb: %w", err)
	}
	if err := raw.Ping(ctx, nil); err != nil {
		_ = raw.Disconnect(context.Background())
		return nil, fmt.Errorf("ping mongodb: %w", err)
	}

	slog.Info("mongodb connected", "addrs", cfg.Addrs, "database", cfg.Database)
	return &Client{
		raw: raw,
		db:  raw.Database(cfg.Database),
	}, nil
}

func (c *Client) Raw() *drivermongo.Client {
	return c.raw
}

func (c *Client) Database() *drivermongo.Database {
	return c.db
}

func (c *Client) Collection(name string) *drivermongo.Collection {
	return c.db.Collection(name)
}

func (c *Client) Ping(ctx context.Context) error {
	if err := c.raw.Ping(ctx, nil); err != nil {
		return fmt.Errorf("ping mongodb: %w", err)
	}
	return nil
}

func (c *Client) Close(ctx context.Context) error {
	if err := c.raw.Disconnect(ctx); err != nil {
		return fmt.Errorf("disconnect mongodb: %w", err)
	}
	return nil
}
