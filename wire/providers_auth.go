package wire

import (
	"house-manager/internal/config"
	v1handler "house-manager/internal/handler/v1"
	"house-manager/internal/repository"
	"house-manager/internal/service"
	authsvc "house-manager/internal/service/auth"
	dbmongo "house-manager/pkg/database/mongo"

	"github.com/google/wire"
)

func newUserRepository(client *dbmongo.Client) *repository.UserRepository {
	return repository.NewUserRepository(client)
}

func newUserAuthRepository(client *dbmongo.Client) *repository.UserAuthRepository {
	return repository.NewUserAuthRepository(client)
}

func newUserProfileExtRepository(client *dbmongo.Client) *repository.UserProfileExtRepository {
	return repository.NewUserProfileExtRepository(client)
}

func newWechatClient(cfg *config.Config) (*authsvc.WechatClient, error) {
	return authsvc.NewWechatClient(authsvc.WechatConfig{
		AppID:   cfg.Wechat.AppID,
		Secret:  cfg.Wechat.Secret,
		APIBase: cfg.Wechat.APIBase,
	})
}

var AuthSet = wire.NewSet(
	newUserRepository,
	newUserAuthRepository,
	newUserProfileExtRepository,
	newWechatClient,
	service.NewAuthService,
	v1handler.NewAuthHandler,
	v1handler.NewSessionHandler,
)
