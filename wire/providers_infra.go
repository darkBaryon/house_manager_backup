package wire

import (
	"context"
	"time"

	"house-manager/internal/config"
	"house-manager/internal/middleware"
	"house-manager/pkg/cache"
	dbmongo "house-manager/pkg/database/mongo"
	dbredis "house-manager/pkg/database/redis"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"golang.org/x/time/rate"
)

func newMongoClient(ctx context.Context, cfg *config.Config) (*dbmongo.Client, error) {
	connectTimeout, err := cfg.MongoDB.ConnectTimeoutDuration()
	if err != nil {
		return nil, err
	}
	socketTimeout, err := cfg.MongoDB.SocketTimeoutDuration()
	if err != nil {
		return nil, err
	}
	serverSelectionTimeout, err := cfg.MongoDB.ServerSelectionTimeoutDuration()
	if err != nil {
		return nil, err
	}

	return dbmongo.NewClient(ctx, dbmongo.Config{
		Addrs:                  cfg.MongoDB.Addrs,
		Database:               cfg.MongoDB.Database,
		AuthSource:             cfg.MongoDB.AuthSource,
		Username:               cfg.MongoDB.Username,
		Password:               cfg.MongoDB.Password,
		PoolSize:               cfg.MongoDB.PoolSize,
		MinPoolSize:            cfg.MongoDB.MinPoolSize,
		ConnectTimeout:         connectTimeout,
		SocketTimeout:          socketTimeout,
		ServerSelectionTimeout: serverSelectionTimeout,
		EnableRetryReads:       cfg.MongoDB.RetryReads,
		EnableRetryWrites:      cfg.MongoDB.RetryWrites,
		ReplicaSet:             cfg.MongoDB.ReplicaSet,
	})
}

func newRedisClient(ctx context.Context, cfg *config.Config) (*dbredis.Client, error) {
	connTimeout, err := cfg.Redis.ConnTimeoutDuration()
	if err != nil {
		return nil, err
	}
	readTimeout, err := cfg.Redis.ReadTimeoutDuration()
	if err != nil {
		return nil, err
	}
	writeTimeout, err := cfg.Redis.WriteTimeoutDuration()
	if err != nil {
		return nil, err
	}
	return dbredis.NewClient(ctx, dbredis.Config{
		Addrs:        cfg.Redis.Addrs,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		ConnTimeout:  connTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		MaxRetries:   cfg.Redis.MaxRetries,
		ClusterMode:  cfg.Redis.ClusterMode,
	})
}

func newEngine(cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	engine := gin.New()
	engine.Use(middleware.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.RateLimit(rate.Every(time.Second), 100))
	return engine
}

func newCache(rc *dbredis.Client) *cache.Cache {
	return cache.New(rc)
}

func newSessionStore(cfg *config.Config, rc *dbredis.Client) (*session.Store, error) {
	ttl, err := cfg.Auth.SessionTTLDuration()
	if err != nil {
		return nil, err
	}
	return session.NewStore(rc, ttl), nil
}

func newContext() context.Context {
	return context.Background()
}

var InfraSet = wire.NewSet(
	newContext,
	newMongoClient,
	newRedisClient,
	newEngine,
	newCache,
	newSessionStore,
)
