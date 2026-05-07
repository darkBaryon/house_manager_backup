package mongo

import (
	"context"
	"fmt"
	"log/slog"

	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Client struct {
	raw *drivermongo.Client
	db  *drivermongo.Database
}

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate mongo config: %w", err)
	}
	if ctx == nil {
		return nil, fmt.Errorf("new mongo client: ctx is nil")
	}

	uri, err := cfg.MongoURI()
	if err != nil {
		return nil, fmt.Errorf("build mongo uri: %w", err)
	}

	opts := options.Client().ApplyURI(uri)

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
		opts.SetConnectTimeout(cfg.ConnectTimeout)
	}
	// Driver v2 does not expose a dedicated socket timeout option; map socket_timeout
	// to the client-level operation timeout so the YAML field remains effective.
	if cfg.SocketTimeout > 0 {
		opts.SetTimeout(cfg.SocketTimeout)
	}
	if cfg.ServerSelectionTimeout > 0 {
		opts.SetServerSelectionTimeout(cfg.ServerSelectionTimeout)
	}
	if cfg.EnableRetryReads {
		opts.SetRetryReads(true)
	}
	if cfg.EnableRetryWrites {
		opts.SetRetryWrites(true)
	}

	raw, err := drivermongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("connect mongodb (%s): %w", cfg.Database, err)
	}
	if err := raw.Ping(ctx, nil); err != nil {
		_ = raw.Disconnect(context.Background())
		return nil, fmt.Errorf("ping mongodb (%s): %w", cfg.Database, err)
	}

	slog.Info("mongodb connected", "addrs", cfg.Addrs, "database", cfg.Database)
	return &Client{
		raw: raw,
		db:  raw.Database(cfg.Database),
	}, nil
}

func (c *Client) Raw() *drivermongo.Client {
	if c == nil {
		return nil
	}
	return c.raw
}

func (c *Client) Database() *drivermongo.Database {
	if c == nil {
		return nil
	}
	return c.db
}

func (c *Client) Collection(name string) *drivermongo.Collection {
	if c == nil || c.db == nil || name == "" {
		return nil
	}
	return c.db.Collection(name)
}

func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.raw == nil {
		return fmt.Errorf("ping mongodb: client is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := c.raw.Ping(ctx, nil); err != nil {
		return fmt.Errorf("ping mongodb database %q: %w", c.dbName(), err)
	}
	return nil
}

func (c *Client) Close(ctx context.Context) error {
	if c == nil || c.raw == nil {
		return fmt.Errorf("disconnect mongodb: client is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := c.raw.Disconnect(ctx); err != nil {
		return fmt.Errorf("disconnect mongodb database %q: %w", c.dbName(), err)
	}
	return nil
}

func (c *Client) dbName() string {
	if c == nil || c.db == nil {
		return ""
	}
	return c.db.Name()
}
