package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

type Config struct {
	Level     string
	Format    string
	AddSource bool
	Service   string
	Env       string
	Fields    map[string]any
}

func ParseLevel(level string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level: %q", level)
	}
}

func ParseFormat(format string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", string(FormatText):
		return FormatText, nil
	case string(FormatJSON):
		return FormatJSON, nil
	default:
		return FormatText, fmt.Errorf("invalid log format: %q", format)
	}
}

// New 构造 logger。wraps 为可选的通用 handler 装饰器，按序包裹基础 handler
// （由 main.go 组装点传入，例如 applog.ContextHandler；本包不依赖具体实现）。
func New(cfg Config, w io.Writer, wraps ...func(slog.Handler) slog.Handler) (*slog.Logger, error) {
	level, err := ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	format, err := ParseFormat(cfg.Format)
	if err != nil {
		return nil, err
	}

	if w == nil {
		w = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	switch format {
	case FormatJSON:
		handler = slog.NewJSONHandler(w, opts)
	case FormatText:
		handler = newPrettyTextHandler(w, opts)
	default:
		return nil, fmt.Errorf("unsupported log format: %q", format)
	}

	for _, wrap := range wraps {
		if wrap != nil {
			handler = wrap(handler)
		}
	}

	logger := slog.New(handler)

	attrs := make([]any, 0, len(cfg.Fields)+2)
	if cfg.Service != "" {
		attrs = append(attrs, "service", cfg.Service)
	}
	if cfg.Env != "" {
		attrs = append(attrs, "env", cfg.Env)
	}
	for k, v := range cfg.Fields {
		attrs = append(attrs, k, v)
	}
	if len(attrs) > 0 {
		logger = logger.With(attrs...)
	}

	return logger, nil
}

func MustNew(cfg Config, w io.Writer, wraps ...func(slog.Handler) slog.Handler) *slog.Logger {
	l, err := New(cfg, w, wraps...)
	if err != nil {
		panic(err)
	}
	return l
}

func Init(cfg Config, w io.Writer, wraps ...func(slog.Handler) slog.Handler) error {
	l, err := New(cfg, w, wraps...)
	if err != nil {
		return err
	}
	slog.SetDefault(l)
	return nil
}

func MustInit(cfg Config, w io.Writer, wraps ...func(slog.Handler) slog.Handler) {
	if err := Init(cfg, w, wraps...); err != nil {
		panic(err)
	}
}
