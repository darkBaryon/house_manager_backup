package wire

import (
	"context"

	v1handler "house-manager/internal/handler/v1"
	dbmongo "house-manager/pkg/database/mongo"
	dbredis "house-manager/pkg/database/redis"

	"github.com/google/wire"
)

func newHealthHandler(mc *dbmongo.Client, rc *dbredis.Client) *v1handler.HealthHandler {
	return &v1handler.HealthHandler{
		MongoPing: func() error { return mc.Ping(context.Background()) },
		RedisPing: func() error { return rc.Ping(context.Background()) },
	}
}

var HealthSet = wire.NewSet(
	newHealthHandler,
)
