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
	return dbmongo.NewClient(ctx, cfg.MongoDB)
}

func newRedisClient(ctx context.Context, cfg *config.Config) (*dbredis.Client, error) {
	return dbredis.NewClient(ctx, cfg.Redis)
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

func newSessionStore(rc *dbredis.Client) *session.Store {
	return session.NewStore(rc, 30*time.Minute)
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
