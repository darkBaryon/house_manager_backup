package wire

import (
	"house-manager/internal/config"
	authdomain "house-manager/internal/domain/auth"
	miniauthhandler "house-manager/internal/handler/v1/miniapp/auth"
	repoauth "house-manager/internal/repository/auth"
	miniauthsvc "house-manager/internal/service/miniapp/auth"
	dbmongo "house-manager/pkg/database/mongo"

	"github.com/google/wire"
)

func newUserRepository(client *dbmongo.Client) *repoauth.UserRepository {
	return repoauth.NewUserRepository(client)
}

func newUserAuthRepository(client *dbmongo.Client) *repoauth.UserAuthRepository {
	return repoauth.NewUserAuthRepository(client)
}

func newUserProfileExtRepository(client *dbmongo.Client) *repoauth.UserProfileExtRepository {
	return repoauth.NewUserProfileExtRepository(client)
}

func newWechatClient(cfg *config.Config) (*authdomain.WechatClient, error) {
	return authdomain.NewWechatClient(authdomain.WechatConfig{
		AppID:   cfg.Wechat.AppID,
		Secret:  cfg.Wechat.Secret,
		APIBase: cfg.Wechat.APIBase,
	})
}

var DomainAuthSet = wire.NewSet(
	newUserRepository,
	newUserAuthRepository,
	newUserProfileExtRepository,
	newWechatClient,
	authdomain.NewUserProfileService,
	authdomain.NewIdentityService,
)

var MiniappAuthSet = wire.NewSet(
	DomainAuthSet,
	miniauthsvc.NewService,
	miniauthhandler.NewAuthHandler,
	miniauthhandler.NewSessionHandler,
)
