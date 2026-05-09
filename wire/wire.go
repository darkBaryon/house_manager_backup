//go:build wireinject

package wire

import (
	"house-manager/internal/app"
	"house-manager/internal/config"

	"github.com/google/wire"
)

func InitializeApp(cfg *config.Config) (*app.App, func(), error) {
	wire.Build(
		InfraSet,
		AuthSet,
		PublishSet,
		HealthSet,
		RouterSet,
		app.NewApp,
	)
	return nil, nil, nil
}
