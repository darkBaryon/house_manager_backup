// Package applog 提供全项目统一的业务日志约定（日志系统收敛方案 v2）：
//   - ContextHandler：在 slog handler 层从 ctx 注入请求级字段（request_id、principal），
//     调用点只需使用标准 slog.XxxContext(ctx, event, kv...)；
//   - Result：错误→日志级别路由的唯一实现（errcode Code<50000 视为预期内失败记 Warn）；
//   - PrincipalGroup：Principal→日志字段映射的唯一实现；
//   - Phone / MaskPhone：手机号脱敏的唯一实现。
package applog

import (
	"context"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"house-manager/pkg/errcode"
	"house-manager/pkg/requestlog"
	"house-manager/pkg/session"
)

// warnCodeLimit 以下的业务错误码视为预期内失败，记 Warn；其余记 Error。
const warnCodeLimit = 50000

// ContextHandler 包装 inner handler，在 Handle 时把 ctx 中的请求级信息
// （request_id、principal）注入为日志字段；无对应 ctx 值时零注入。
// 由 main.go 组装点经 logger.Init 的 wraps 参数接入，pkg/logger 不依赖本包。
//
// 注意：注入的字段会落在已通过 WithGroup 打开的 group 之内。本仓库业务日志
// 不使用 WithGroup，故按顶层字段处理；如未来引入 WithGroup，需要在此记录
// group 栈并将请求级字段写到顶层。
func ContextHandler(inner slog.Handler) slog.Handler {
	return &ctxHandler{inner: inner}
}

type ctxHandler struct {
	inner slog.Handler
}

func (h *ctxHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *ctxHandler) Handle(ctx context.Context, record slog.Record) error {
	if requestID := requestlog.RequestIDFromContext(ctx); requestID != "" {
		record.AddAttrs(slog.String("request_id", requestID))
	}
	if principal, ok := session.PrincipalFromContext(ctx); ok {
		record.AddAttrs(PrincipalGroup(principal))
	}
	return h.inner.Handle(ctx, record)
}

func (h *ctxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ctxHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *ctxHandler) WithGroup(name string) slog.Handler {
	return &ctxHandler{inner: h.inner.WithGroup(name)}
}

// PrincipalGroup 是全项目唯一的 Principal→日志字段映射，
// ContextHandler 与 middleware 访问日志共用。
func PrincipalGroup(principal session.Principal) slog.Attr {
	args := make([]any, 0, 4)
	args = append(args,
		slog.String("terminal", principal.Terminal),
		slog.String("type", principal.PrincipalType),
		slog.String("id", principal.PrincipalID),
	)
	if principal.Phone != "" {
		args = append(args, slog.Any("phone", Phone(principal.Phone)))
	}
	return slog.Group("principal", args...)
}

// Result 按操作结果落日志：err 为 nil 记 Info(successEvent)；
// errcode Code<50000 记 Warn(failureEvent)；其余记 Error(failureEvent)。
// 失败路径自动追加 "error" 字段（行为保持自原 log*Result）。
func Result(ctx context.Context, successEvent, failureEvent string, err error, attrs ...any) {
	if err == nil {
		emit(ctx, slog.LevelInfo, successEvent, attrs)
		return
	}
	attrs = append(attrs, "error", err)
	if code := errcode.FromError(err); code != nil && code.Code < warnCodeLimit {
		emit(ctx, slog.LevelWarn, failureEvent, attrs)
		return
	}
	emit(ctx, slog.LevelError, failureEvent, attrs)
}

// emit 手工构造 Record 以使 source 指向 Result 的业务调用点。
// 这是全项目唯一允许使用 runtime.Callers 的位置（方案「禁止事项」）。
func emit(ctx context.Context, level slog.Level, event string, attrs []any) {
	logger := slog.Default()
	if !logger.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // 跳过 Callers、emit、Result，定位业务调用点
	record := slog.NewRecord(time.Now(), level, event, pcs[0])
	record.Add(attrs...)
	_ = logger.Handler().Handle(ctx, record)
}

// Phone 在日志中自动脱敏的手机号类型。
// 用法：slog.InfoContext(ctx, event, "phone", applog.Phone(raw))。
type Phone string

// LogValue 实现 slog.LogValuer，输出脱敏后的号码。
func (p Phone) LogValue() slog.Value {
	return slog.StringValue(MaskPhone(string(p)))
}

// MaskPhone 手机号脱敏的唯一实现（含 TrimSpace，语义采纳自原 middleware 版本）。
func MaskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
