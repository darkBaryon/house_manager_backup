package wire

import (
	"context"
	"time"

	"house-manager/internal/config"
	"house-manager/internal/middleware"
	"house-manager/pkg/cache"
	"house-manager/pkg/database"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"golang.org/x/time/rate"
)

func newMongoConfig(cfg *config.Config) database.MongoConfig {
	return database.MongoConfig{
		Addrs:                  cfg.MongoDB.Addrs,
		Database:               cfg.MongoDB.Database,
		AuthSource:             cfg.MongoDB.AuthSource,
		Username:               cfg.MongoDB.Username,
		Password:               cfg.MongoDB.Password,
		PoolSize:               cfg.MongoDB.PoolSize,
		MinPoolSize:            cfg.MongoDB.MinPoolSize,
		ConnectTimeout:         cfg.MongoDB.ConnectTimeout,
		SocketTimeout:          cfg.MongoDB.SocketTimeout,
		ServerSelectionTimeout: cfg.MongoDB.ServerSelectionTimeout,
		MaxRetries:             cfg.MongoDB.MaxRetries,
		ReplicaSet:             cfg.MongoDB.ReplicaSet,
	}
}

func newRedisConfig(cfg *config.Config) database.RedisConfig {
	return database.RedisConfig{
		Addrs:        cfg.Redis.Addrs,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		ConnTimeout:  cfg.Redis.ConnTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		MaxRetries:   cfg.Redis.MaxRetries,
		ClusterMode:  cfg.Redis.ClusterMode,
	}
}

func newEngine(cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	engine := gin.New()
	engine.Use(middleware.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.RateLimit(rate.Every(time.Second), 100))
	return engine
}

func newCache(rc database.RedisClient) *cache.Cache {
	return cache.New(cache.NewRedisAdapter(rc))
}

func newSessionStore(rc database.RedisClient) *session.Store {
	return session.NewStore(cache.NewRedisAdapter(rc), 30*time.Minute)
}

func newContext() context.Context {
	return context.Background()
}

var InfraSet = wire.NewSet(
	newContext,
	newMongoConfig,
	newRedisConfig,
	database.NewMongoClient,
	database.NewRedisClient,
	newEngine,
	newSessionStore,
)
