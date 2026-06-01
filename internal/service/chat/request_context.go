package chat

import (
	"context"
	"strings"

	"house-manager/pkg/requestlog"
)

func requestIDFromContext(ctx context.Context) string {
	return strings.TrimSpace(requestlog.RequestIDFromContext(ctx))
}
