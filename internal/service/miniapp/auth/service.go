package auth

import (
	"context"
	"errors"
	"house-manager/pkg/applog"
	"log/slog"

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
	slog.InfoContext(ctx, "miniapp.auth.login.start", "login_ip", loginIP,
		"code_present", code != "",
	)
	userID, err := s.identity.WechatLogin(ctx, code, loginIP)
	if err != nil {
		slog.WarnContext(ctx, "miniapp.auth.login.failed", "login_ip", loginIP,
			"error", err,
		)
		return "", err
	}
	token, err := s.issueTokenForUser(ctx, userID, "登录失败，请稍后重试")
	if err != nil {
		slog.WarnContext(ctx, "miniapp.auth.login.failed", "user_id", userID.Hex(),
			"login_ip", loginIP,
			"error", err,
		)
		return "", err
	}
	slog.InfoContext(ctx, "miniapp.auth.login.success", "user_id", userID.Hex(),
		"login_ip", loginIP,
		"token_issued", token != "",
	)
	return token, nil
}

func (s *Service) WechatRegister(ctx context.Context, code, phoneCode, loginIP string) (string, error) {
	slog.InfoContext(ctx, "miniapp.auth.register.start", "login_ip", loginIP,
		"code_present", code != "",
		"phone_code_present", phoneCode != "",
	)
	userID, err := s.identity.WechatRegister(ctx, code, phoneCode, loginIP)
	if err != nil {
		slog.WarnContext(ctx, "miniapp.auth.register.failed", "login_ip", loginIP,
			"error", err,
		)
		return "", err
	}
	token, err := s.issueTokenForUser(ctx, userID, "注册失败，请稍后重试")
	if err != nil {
		slog.WarnContext(ctx, "miniapp.auth.register.failed", "user_id", userID.Hex(),
			"login_ip", loginIP,
			"error", err,
		)
		return "", err
	}
	slog.InfoContext(ctx, "miniapp.auth.register.success", "user_id", userID.Hex(),
		"login_ip", loginIP,
		"token_issued", token != "",
	)
	return token, nil
}

func (s *Service) issueTokenForUser(ctx context.Context, userID bson.ObjectID, loginFailedDetail string) (string, error) {
	if userID.IsZero() {
		return "", errcode.InvalidParam.WithErrorf("userID is required")
	}
	slog.InfoContext(ctx, "miniapp.auth.issue_token.start", "user_id", userID.Hex())
	if err := s.userProfile.TouchUserLastActive(ctx, userID); err != nil {
		slog.WarnContext(ctx, "miniapp.auth.issue_token.failed", "user_id", userID.Hex(),
			"step", "touch_last_active",
			"error", err,
		)
		return "", err
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.issue_token.failed", "user_id", userID.Hex(),
			"step", "find_user",
			"error", err,
		)
		return "", errcode.DatabaseError.WithError(err)
	}
	if user == nil {
		slog.ErrorContext(ctx, "miniapp.auth.issue_token.failed", "user_id", userID.Hex(),
			"step", "find_user",
			"error", "user not found",
		)
		return "", errcode.DatabaseError.WithError(errors.New("user not found"))
	}
	token, err := s.sessionStore.CreateMiniappUser(ctx, userID.Hex(), user.Phone)
	if err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.issue_token.failed", "user_id", userID.Hex(),
			"phone", applog.Phone(user.Phone),
			"step", "create_session",
			"error", err,
		)
		return "", errcode.CacheError.WithError(err)
	}
	if token == "" {
		slog.ErrorContext(ctx, "miniapp.auth.issue_token.failed", "user_id", userID.Hex(),
			"phone", applog.Phone(user.Phone),
			"step", "create_session",
			"error", loginFailedDetail,
		)
		return "", errcode.InternalError.WithError(errors.New(loginFailedDetail))
	}
	slog.InfoContext(ctx, "miniapp.auth.issue_token.success", "user_id", userID.Hex(),
		"phone", applog.Phone(user.Phone),
	)
	return token, nil
}
