package wire

import (
	"context"

	v1handler "house-manager/internal/handler/v1"
	"house-manager/pkg/database"

	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func newHealthHandler(mc *mongo.Client, rc database.RedisClient) *v1handler.HealthHandler {
	return &v1handler.HealthHandler{
		MongoPing: func() error { return mc.Ping(context.Background(), nil) },
		RedisPing: func() error { return rc.Ping(context.Background()).Err() },
	}
}

var HealthSet = wire.NewSet(
	newHealthHandler,
)
