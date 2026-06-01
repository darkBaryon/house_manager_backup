package logging

import (
	"context"
	"strings"

	"house-manager/pkg/requestlog"
	"house-manager/pkg/session"
)

func Attrs(ctx context.Context, attrs ...any) []any {
	base := make([]any, 0, len(attrs)+8)
	if requestID := requestlog.RequestIDFromContext(ctx); requestID != "" {
		base = append(base, "request_id", requestID)
	}
	if principal, ok := session.PrincipalFromContext(ctx); ok {
		base = append(base, "principal", CompactPrincipal(principal))
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

func CompactPrincipal(principal session.Principal) string {
	parts := []string{
		strings.TrimSpace(principal.Terminal),
		strings.TrimSpace(principal.PrincipalType),
	}
	if phone := MaskPhone(principal.Phone); phone != "" {
		parts = append(parts, phone)
	} else if principal.PrincipalID != "" {
		parts = append(parts, strings.TrimSpace(principal.PrincipalID))
	}
	return strings.Join(parts, ":")
}
