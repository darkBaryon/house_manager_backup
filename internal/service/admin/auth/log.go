package auth

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"house-manager/pkg/errcode"
	"house-manager/pkg/requestlog"
	"house-manager/pkg/session"
)

func logAuthInfo(ctx context.Context, event string, attrs ...any) {
	emitAuthLog(ctx, slog.LevelInfo, event, attrs...)
}

func logAuthWarn(ctx context.Context, event string, attrs ...any) {
	emitAuthLog(ctx, slog.LevelWarn, event, attrs...)
}

func logAuthError(ctx context.Context, event string, attrs ...any) {
	emitAuthLog(ctx, slog.LevelError, event, attrs...)
}

func logAuthResult(ctx context.Context, successEvent string, failureEvent string, err error, attrs ...any) {
	if err == nil {
		logAuthInfo(ctx, successEvent, attrs...)
		return
	}
	errorAttrs := append(attrs, "error", err)
	code := errcode.FromError(err)
	if code != nil && code.Code < 50000 {
		logAuthWarn(ctx, failureEvent, errorAttrs...)
		return
	}
	logAuthError(ctx, failureEvent, errorAttrs...)
}

func authLogAttrs(ctx context.Context, attrs ...any) []any {
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
		if principal.Phone != "" {
			base = append(base, "principal_phone", maskPhone(principal.Phone))
		}
	}
	base = append(base, attrs...)
	return base
}

func emitAuthLog(ctx context.Context, level slog.Level, event string, attrs ...any) {
	logger := slog.Default()
	if !logger.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	record := slog.NewRecord(time.Now(), level, event, pcs[0])
	record.AddAttrs(pairsToAttrs(authLogAttrs(ctx, attrs...))...)
	_ = logger.Handler().Handle(ctx, record)
}

func pairsToAttrs(values []any) []slog.Attr {
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

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
