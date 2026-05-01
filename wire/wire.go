//go:build wireinject

package wire

import (
	"house-manager/internal/app"
	"house-manager/internal/config"

	"github.com/google/wire"
)

func InitializeApp(cfgPath string) (*app.App, func(), error) {
	wire.Build(
		config.Load,
		InfraSet,
		RoomSet,
		HealthSet,
		newRouteGroups,
		app.NewApp,
	)
	return nil, nil, nil
}
