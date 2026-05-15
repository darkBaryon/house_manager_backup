package auth

import (
	"context"
	"errors"

	miniappauthrepo "house-manager/internal/repository/miniapp_auth"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Service 是小程序端认证应用服务入口。
type Service struct {
	identity     *IdentityService
	userProfile  *UserProfileService
	userRepo     *miniappauthrepo.UserRepository
	sessionStore *session.Store
}

func NewService(
	identity *IdentityService,
	userProfile *UserProfileService,
	userRepo *miniappauthrepo.UserRepository,
	sessionStore *session.Store,
) *Service {
	return &Service{
		identity:     identity,
		userProfile:  userProfile,
		userRepo:     userRepo,
		sessionStore: sessionStore,
	}
}

func (s *Service) WechatLogin(ctx context.Context, code, loginIP string) (string, error) {
	userID, err := s.identity.WechatLogin(ctx, code, loginIP)
	if err != nil {
		return "", err
	}
	return s.issueTokenForUser(ctx, userID, "登录失败，请稍后重试")
}

func (s *Service) WechatRegister(ctx context.Context, code, phoneCode, loginIP string) (string, error) {
	userID, err := s.identity.WechatRegister(ctx, code, phoneCode, loginIP)
	if err != nil {
		return "", err
	}
	return s.issueTokenForUser(ctx, userID, "注册失败，请稍后重试")
}

func (s *Service) issueTokenForUser(ctx context.Context, userID bson.ObjectID, loginFailedDetail string) (string, error) {
	if userID.IsZero() {
		return "", errcode.InvalidParam.WithErrorf("userID is required")
	}
	if err := s.userProfile.TouchUserLastActive(ctx, userID); err != nil {
		return "", err
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", errcode.DatabaseError.WithError(err)
	}
	if user == nil {
		return "", errcode.DatabaseError.WithError(errors.New("user not found"))
	}
	token, err := s.sessionStore.CreateMiniappUser(ctx, userID.Hex(), user.Phone)
	if err != nil {
		return "", errcode.CacheError.WithError(err)
	}
	if token == "" {
		return "", errcode.InternalError.WithError(errors.New(loginFailedDetail))
	}
	return token, nil
}
