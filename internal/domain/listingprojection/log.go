package listingprojection

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"house-manager/pkg/requestlog"
	"house-manager/pkg/session"
)

func logProjectionInfo(ctx context.Context, event string, attrs ...any) {
	emitProjectionLog(ctx, slog.LevelInfo, event, attrs...)
}

func logProjectionWarn(ctx context.Context, event string, attrs ...any) {
	emitProjectionLog(ctx, slog.LevelWarn, event, attrs...)
}

func logProjectionError(ctx context.Context, event string, attrs ...any) {
	emitProjectionLog(ctx, slog.LevelError, event, attrs...)
}

func emitProjectionLog(ctx context.Context, level slog.Level, event string, attrs ...any) {
	logger := slog.Default()
	if !logger.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	record := slog.NewRecord(time.Now(), level, event, pcs[0])
	record.AddAttrs(projectionAttrs(projectionLogAttrs(ctx, attrs...))...)
	_ = logger.Handler().Handle(ctx, record)
}

func projectionLogAttrs(ctx context.Context, attrs ...any) []any {
	base := make([]any, 0, len(attrs)+4)
	if requestID := requestlog.RequestIDFromContext(ctx); requestID != "" {
		base = append(base, "request_id", requestID)
	}
	if principal, ok := session.PrincipalFromContext(ctx); ok {
		base = append(base,
			"terminal", principal.Terminal,
			"principal_type", principal.PrincipalType,
			"principal_id", principal.PrincipalID,
		)
	}
	base = append(base, attrs...)
	return base
}

func projectionAttrs(values []any) []slog.Attr {
	if len(values) == 0 {
		return nil
	}
	attrs := make([]slog.Attr, 0, len(values)/2)
	for i := 0; i+1 < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok || key == "" {
			continue
		}
		attrs = append(attrs, slog.Any(key, values[i+1]))
	}
	return attrs
}
