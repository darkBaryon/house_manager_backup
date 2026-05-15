package wire

import (
	"house-manager/internal/config"
	miniauthhandler "house-manager/internal/handler/v1/miniapp/auth"
	miniappauthrepo "house-manager/internal/repository/miniapp_auth"
	miniauthsvc "house-manager/internal/service/miniapp/auth"
	dbmongo "house-manager/pkg/database/mongo"

	"github.com/google/wire"
)

func newUserRepository(client *dbmongo.Client) *miniappauthrepo.UserRepository {
	return miniappauthrepo.NewUserRepository(client)
}

func newUserAuthRepository(client *dbmongo.Client) *miniappauthrepo.UserAuthRepository {
	return miniappauthrepo.NewUserAuthRepository(client)
}

func newUserProfileExtRepository(client *dbmongo.Client) *miniappauthrepo.UserProfileExtRepository {
	return miniappauthrepo.NewUserProfileExtRepository(client)
}

func newWechatClient(cfg *config.Config) (*miniauthsvc.WechatClient, error) {
	return miniauthsvc.NewWechatClient(miniauthsvc.WechatConfig{
		AppID:   cfg.Wechat.AppID,
		Secret:  cfg.Wechat.Secret,
		APIBase: cfg.Wechat.APIBase,
	})
}

var MiniappAuthDomainSet = wire.NewSet(
	newUserRepository,
	newUserAuthRepository,
	newUserProfileExtRepository,
	newWechatClient,
	miniauthsvc.NewUserProfileService,
	miniauthsvc.NewIdentityService,
)

var MiniappAuthSet = wire.NewSet(
	MiniappAuthDomainSet,
	miniauthsvc.NewService,
	miniauthhandler.NewAuthHandler,
	miniauthhandler.NewSessionHandler,
)
