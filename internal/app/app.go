package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"house-manager/internal/config"
	"house-manager/internal/handler"
	dbmongo "house-manager/pkg/database/mongo"
	dbredis "house-manager/pkg/database/redis"

	"github.com/gin-gonic/gin"
)

// RouteGroup 路由组（前缀 + 中间件 + handler 注册器）
type RouteGroup struct {
	Prefix     string
	Middleware []gin.HandlerFunc
	Registrars []handler.RouteRegistrar
}

// App 应用实例
type App struct {
	Config      *config.Config
	Engine      *gin.Engine
	Server      *http.Server
	MongoClient *dbmongo.Client
	RedisClient *dbredis.Client
}

// NewApp 创建应用实例（依赖由 Wire 注入）
func NewApp(
	cfg *config.Config,
	engine *gin.Engine,
	mongoClient *dbmongo.Client,
	redisClient *dbredis.Client,
	groups []RouteGroup,
) (*App, func(), error) {
	a := &App{
		Config:      cfg,
		Engine:      engine,
		MongoClient: mongoClient,
		RedisClient: redisClient,
	}

	for _, g := range groups {
		rg := engine.Group(g.Prefix, g.Middleware...)
		for _, r := range g.Registrars {
			r.RegisterRoutes(rg)
		}
	}

	cleanup := func() {
		a.Close()
	}

	return a, cleanup, nil
}

// Shutdown 优雅关闭 HTTP Server，等待活跃请求完成
func (a *App) Shutdown(ctx context.Context) error {
	if a.Server != nil {
		slog.Info("shutting down http server...")
		return a.Server.Shutdown(ctx)
	}
	return nil
}

// Close 关闭基础设施连接（MongoDB、Redis）
func (a *App) Close() {
	if a.MongoClient != nil {
		if err := a.MongoClient.Close(context.Background()); err != nil {
			slog.Error("mongodb disconnect error", "error", err)
		} else {
			slog.Info("mongodb disconnected")
		}
	}
	if a.RedisClient != nil {
		if err := a.RedisClient.Close(); err != nil {
			slog.Error("redis close error", "error", err)
		} else {
			slog.Info("redis closed")
		}
	}
}

// Start 启动 HTTP Server
func (a *App) Start() error {
	addr := fmt.Sprintf(":%d", a.Config.Server.Port)
	a.Server = &http.Server{
		Addr:    addr,
		Handler: a.Engine,
	}
	slog.Info("starting server", "addr", addr)
	return a.Server.ListenAndServe()
}
