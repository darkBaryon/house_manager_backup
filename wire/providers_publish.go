package wire

import (
	publishhandler "house-manager/internal/handler/v1/publish"
	publishsvc "house-manager/internal/service/publish"

	"github.com/google/wire"
)

var PublishSet = wire.NewSet(
	publishsvc.NewPublishService,
	publishhandler.NewPublishHandler,
)
