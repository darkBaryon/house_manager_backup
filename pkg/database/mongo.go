package database

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoConfig struct {
	Addrs                 []string
	Database              string
	AuthSource            string
	Username              string
	Password              string
	PoolSize              uint64
	MinPoolSize           uint64
	ConnectTimeout        int
	SocketTimeout         int
	ServerSelectionTimeout int
	MaxRetries            int
	ReplicaSet            string
}

func (m MongoConfig) BuildURI() string {
	return "mongodb://" + strings.Join(m.Addrs, ",")
}

func NewMongoClient(ctx context.Context, cfg MongoConfig) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(cfg.BuildURI())

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

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect mongodb: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	slog.Info("mongodb connected", "addrs", cfg.Addrs, "database", cfg.Database)
	return client, nil
}

func GetDatabase(client *mongo.Client, dbName string) *mongo.Database {
	return client.Database(dbName)
}
