package logging

import (
	"context"

	"house-manager/pkg/requestlog"
	"house-manager/pkg/session"
)

func Attrs(ctx context.Context, attrs ...any) []any {
	base := make([]any, 0, len(attrs)+8)
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
			base = append(base, "principal_phone", MaskPhone(principal.Phone))
		}
	}
	base = append(base, attrs...)
	return base
}

func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
