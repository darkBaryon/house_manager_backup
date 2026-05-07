package service

import (
	"context"
	"errors"
	"fmt"

	"house-manager/internal/repository"
	authsvc "house-manager/internal/service/auth"
	dbmongo "house-manager/pkg/database/mongo"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// AuthService 认证模块对外统一入口。
type AuthService struct {
	identity     *authsvc.IdentityService
	userProfile  *authsvc.UserProfileService
	sessionStore *session.Store
}

func NewAuthService(
	mongoClient *dbmongo.Client,
	userRepo *repository.UserRepository,
	authRepo *repository.UserAuthRepository,
	profileExtRepo *repository.UserProfileExtRepository,
	wechatClient *authsvc.WechatClient,
	sessionStore *session.Store,
) *AuthService {
	userProfile := authsvc.NewUserProfileService(userRepo, profileExtRepo)
	service := &AuthService{
		userProfile:  userProfile,
		sessionStore: sessionStore,
	}
	service.identity = authsvc.NewIdentityService(
		mongoClient,
		authRepo,
		userRepo,
		userProfile,
		wechatClient,
	)
	return service
}

func (s *AuthService) WechatLogin(ctx context.Context, code, loginIP string) (string, error) {
	userID, err := s.identity.WechatLogin(ctx, code, loginIP)
	if err != nil {
		return "", err
	}
	return s.issueTokenForUser(ctx, userID, "登录失败，请稍后重试")
}

func (s *AuthService) WechatRegister(ctx context.Context, code, phoneCode, loginIP string) (string, error) {
	userID, err := s.identity.WechatRegister(ctx, code, phoneCode, loginIP)
	if err != nil {
		return "", err
	}
	return s.issueTokenForUser(ctx, userID, "注册失败，请稍后重试")
}

func (s *AuthService) issueTokenForUser(ctx context.Context, userID bson.ObjectID, loginFailedDetail string) (string, error) {
	if userID.IsZero() {
		return "", errcode.InvalidParam.WithError(fmt.Errorf("userID is required"))
	}
	if err := s.userProfile.TouchUserLastActive(ctx, userID); err != nil {
		return "", err
	}
	token, err := s.sessionStore.Create(ctx, userID.Hex())
	if err != nil {
		return "", errcode.CacheError.WithError(err)
	}
	if token == "" {
		return "", errcode.InternalError.WithError(errors.New(loginFailedDetail))
	}
	return token, nil
}
