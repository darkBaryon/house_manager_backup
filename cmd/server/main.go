package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"house-manager/internal/config"
	"house-manager/pkg/configpath"
	"house-manager/pkg/logger"
	"house-manager/wire"
)

func main() {
	cfgFile := flag.String("c", "", "config file path")
	flag.Parse()

	cfgPath, err := configpath.Resolve(*cfgFile)
	if err != nil {
		slog.Error("failed to resolve config", "error", err)
		os.Exit(1)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if err := logger.Init(logger.Config{
		Level:     cfg.Log.Level,
		Format:    cfg.Log.Format,
		AddSource: cfg.Log.AddSource,
		Service:   cfg.Log.Service,
		Env:       cfg.Log.Env,
		Fields:    cfg.Log.Fields,
	}, os.Stdout); err != nil {
		slog.Error("failed to initialize logger", "error", err)
		os.Exit(1)
	}

	appl, cleanup, err := wire.InitializeApp(cfg)
	if err != nil {
		slog.Error("failed to initialize app", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	slog.Info("infrastructure check passed")

	// 启动 HTTP Server
	go func() {
		if err := appl.Start(); err != nil && err.Error() != "http: Server closed" {
			slog.Error("server error", "error", err)
		}
	}()

	// 等待退出信号
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	slog.Info("shutting down server...")

	// 优雅关闭：等待活跃请求完成（最多 10 秒）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := appl.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	slog.Info("server stopped")
}
