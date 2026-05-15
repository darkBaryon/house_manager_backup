package auth

import (
	"context"
	"fmt"
	authmodel "house-manager/internal/model/auth"
	miniappauthrepo "house-manager/internal/repository/miniapp_auth"
	miniapplog "house-manager/internal/service/miniapp/logging"
	"house-manager/pkg/errcode"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// IdentityService 负责微信身份注册/登录与 user_id 绑定。
type IdentityService struct {
	authRepo     *miniappauthrepo.UserAuthRepository
	userRepo     *miniappauthrepo.UserRepository
	userProfile  *UserProfileService
	wechatClient *WechatClient
}

func NewIdentityService(
	authRepo *miniappauthrepo.UserAuthRepository,
	userRepo *miniappauthrepo.UserRepository,
	userProfile *UserProfileService,
	wechatClient *WechatClient,
) *IdentityService {
	return &IdentityService{
		authRepo:     authRepo,
		userRepo:     userRepo,
		userProfile:  userProfile,
		wechatClient: wechatClient,
	}
}

func (s *IdentityService) WechatLogin(ctx context.Context, code, loginIP string) (bson.ObjectID, error) {
	if code == "" {
		return bson.NilObjectID, errcode.InvalidParam.WithError(fmt.Errorf("code is required"))
	}
	slog.InfoContext(ctx, "miniapp.auth.identity.wechat_login.start", miniapplog.Attrs(ctx,
		"login_ip", loginIP,
	)...)

	openID, _, err := s.wechatClient.ExchangeLoginCode(ctx, code)
	if err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.identity.wechat_login.failed", miniapplog.Attrs(ctx,
			"step", "exchange_login_code",
			"login_ip", loginIP,
			"error", err,
		)...)
		return bson.NilObjectID, errcode.InternalError.WithError(err)
	}

	authRecord, err := s.authRepo.FindByOpenID(ctx, authmodel.AuthProviderWechat, openID)
	if err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.identity.wechat_login.failed", miniapplog.Attrs(ctx,
			"step", "find_auth_by_openid",
			"login_ip", loginIP,
			"error", err,
		)...)
		return bson.NilObjectID, errcode.DatabaseError.WithError(err)
	}
	if authRecord == nil {
		slog.WarnContext(ctx, "miniapp.auth.identity.wechat_login.failed", miniapplog.Attrs(ctx,
			"step", "find_auth_by_openid",
			"login_ip", loginIP,
			"reason", "wechat account is not registered",
		)...)
		return bson.NilObjectID, errcode.Unauthorized.WithError(fmt.Errorf("wechat account is not registered"))
	}

	if err := s.authRepo.TouchLastLogin(ctx, authRecord.ID, time.Now().Unix(), loginIP); err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.identity.wechat_login.failed", miniapplog.Attrs(ctx,
			"step", "touch_last_login",
			"user_id", authRecord.UserID.Hex(),
			"login_ip", loginIP,
			"error", err,
		)...)
		return bson.NilObjectID, errcode.DatabaseError.WithError(err)
	}
	slog.InfoContext(ctx, "miniapp.auth.identity.wechat_login.success", miniapplog.Attrs(ctx,
		"user_id", authRecord.UserID.Hex(),
		"login_ip", loginIP,
	)...)
	return authRecord.UserID, nil
}

func (s *IdentityService) WechatRegister(ctx context.Context, code, phoneCode, loginIP string) (bson.ObjectID, error) {
	if code == "" {
		return bson.NilObjectID, errcode.InvalidParam.WithError(fmt.Errorf("code is required"))
	}
	if phoneCode == "" {
		return bson.NilObjectID, errcode.InvalidParam.WithError(fmt.Errorf("phone code is required"))
	}
	slog.InfoContext(ctx, "miniapp.auth.identity.wechat_register.start", miniapplog.Attrs(ctx,
		"login_ip", loginIP,
	)...)

	openID, unionID, err := s.wechatClient.ExchangeLoginCode(ctx, code)
	if err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.identity.wechat_register.failed", miniapplog.Attrs(ctx,
			"step", "exchange_login_code",
			"login_ip", loginIP,
			"error", err,
		)...)
		return bson.NilObjectID, errcode.InternalError.WithError(err)
	}
	phone, err := s.wechatClient.ExchangePhoneCode(ctx, phoneCode)
	if err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.identity.wechat_register.failed", miniapplog.Attrs(ctx,
			"step", "exchange_phone_code",
			"login_ip", loginIP,
			"error", err,
		)...)
		return bson.NilObjectID, errcode.InternalError.WithError(err)
	}
	slog.InfoContext(ctx, "miniapp.auth.identity.wechat_register.phone_resolved", miniapplog.Attrs(ctx,
		"login_ip", loginIP,
		"phone", miniapplog.MaskPhone(phone),
		"union_id_present", unionID != "",
	)...)

	now := time.Now().Unix()

	wechatAuth, err := s.authRepo.FindByOpenID(ctx, authmodel.AuthProviderWechat, openID)
	if err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.identity.wechat_register.failed", miniapplog.Attrs(ctx,
			"step", "find_auth_by_openid",
			"phone", miniapplog.MaskPhone(phone),
			"error", err,
		)...)
		return bson.NilObjectID, errcode.DatabaseError.WithError(err)
	}
	if wechatAuth != nil {
		slog.InfoContext(ctx, "miniapp.auth.identity.wechat_register.reuse_openid_binding", miniapplog.Attrs(ctx,
			"user_id", wechatAuth.UserID.Hex(),
			"phone", miniapplog.MaskPhone(phone),
		)...)
		if err := s.ensureAuthPhoneMatches(ctx, wechatAuth, phone); err != nil {
			slog.WarnContext(ctx, "miniapp.auth.identity.wechat_register.failed", miniapplog.Attrs(ctx,
				"step", "ensure_auth_phone_matches",
				"user_id", wechatAuth.UserID.Hex(),
				"phone", miniapplog.MaskPhone(phone),
				"error", err,
			)...)
			return bson.NilObjectID, err
		}
		userID, err := s.touchWechatAuth(ctx, wechatAuth, now, loginIP)
		if err != nil {
			slog.WarnContext(ctx, "miniapp.auth.identity.wechat_register.failed", miniapplog.Attrs(ctx,
				"step", "touch_wechat_auth",
				"user_id", wechatAuth.UserID.Hex(),
				"phone", miniapplog.MaskPhone(phone),
				"error", err,
			)...)
			return bson.NilObjectID, err
		}
		slog.InfoContext(ctx, "miniapp.auth.identity.wechat_register.success", miniapplog.Attrs(ctx,
			"user_id", userID.Hex(),
			"phone", miniapplog.MaskPhone(phone),
			"path", "existing_openid_binding",
		)...)
		return userID, nil
	}

	unionAuth, err := s.findWechatAuthByUnionID(ctx, unionID)
	if err != nil {
		return bson.NilObjectID, err
	}
	if unionAuth != nil {
		slog.InfoContext(ctx, "miniapp.auth.identity.wechat_register.reuse_unionid_binding", miniapplog.Attrs(ctx,
			"user_id", unionAuth.UserID.Hex(),
			"phone", miniapplog.MaskPhone(phone),
		)...)
		if err := s.ensureAuthPhoneMatches(ctx, unionAuth, phone); err != nil {
			slog.WarnContext(ctx, "miniapp.auth.identity.wechat_register.failed", miniapplog.Attrs(ctx,
				"step", "ensure_auth_phone_matches",
				"user_id", unionAuth.UserID.Hex(),
				"phone", miniapplog.MaskPhone(phone),
				"error", err,
			)...)
			return bson.NilObjectID, err
		}
		userID, err := s.updateWechatAuthOpenID(ctx, unionAuth, phone, openID, unionID, now, loginIP)
		if err != nil {
			slog.WarnContext(ctx, "miniapp.auth.identity.wechat_register.failed", miniapplog.Attrs(ctx,
				"step", "update_wechat_auth_openid",
				"user_id", unionAuth.UserID.Hex(),
				"phone", miniapplog.MaskPhone(phone),
				"error", err,
			)...)
			return bson.NilObjectID, err
		}
		slog.InfoContext(ctx, "miniapp.auth.identity.wechat_register.success", miniapplog.Attrs(ctx,
			"user_id", userID.Hex(),
			"phone", miniapplog.MaskPhone(phone),
			"path", "existing_unionid_binding",
		)...)
		return userID, nil
	}

	userID, err := s.findOrCreateUserByPhone(ctx, phone)
	if err != nil {
		slog.WarnContext(ctx, "miniapp.auth.identity.wechat_register.failed", miniapplog.Attrs(ctx,
			"step", "find_or_create_user_by_phone",
			"phone", miniapplog.MaskPhone(phone),
			"error", err,
		)...)
		return bson.NilObjectID, err
	}
	slog.InfoContext(ctx, "miniapp.auth.identity.wechat_register.user_resolved", miniapplog.Attrs(ctx,
		"user_id", userID.Hex(),
		"phone", miniapplog.MaskPhone(phone),
	)...)

	boundUserID, err := s.ensureWechatBound(ctx, userID, phone, openID, unionID, now, loginIP)
	if err != nil {
		slog.WarnContext(ctx, "miniapp.auth.identity.wechat_register.failed", miniapplog.Attrs(ctx,
			"step", "ensure_wechat_bound",
			"user_id", userID.Hex(),
			"phone", miniapplog.MaskPhone(phone),
			"error", err,
		)...)
		return bson.NilObjectID, err
	}
	slog.InfoContext(ctx, "miniapp.auth.identity.wechat_register.success", miniapplog.Attrs(ctx,
		"user_id", boundUserID.Hex(),
		"phone", miniapplog.MaskPhone(phone),
		"path", "new_or_existing_phone_user",
	)...)
	return boundUserID, nil
}

func (s *IdentityService) findWechatAuthByUnionID(ctx context.Context, unionID string) (*authmodel.UserAuth, error) {
	if unionID == "" {
		return nil, nil
	}
	authRecord, err := s.authRepo.FindByUnionID(ctx, authmodel.AuthProviderWechat, unionID)
	if err != nil {
		return nil, errcode.DatabaseError.WithError(err)
	}
	return authRecord, nil
}

func (s *IdentityService) findOrCreateUserByPhone(ctx context.Context, phone string) (bson.ObjectID, error) {
	phoneUser, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return bson.NilObjectID, errcode.DatabaseError.WithError(err)
	}
	if phoneUser != nil {
		if phoneUser.ID.IsZero() {
			return bson.NilObjectID, errcode.DatabaseError.WithError(fmt.Errorf("phone user id is required"))
		}
		if err := s.userProfile.EnsureUserProfileExt(ctx, phoneUser.ID); err != nil {
			return bson.NilObjectID, err
		}
		return phoneUser.ID, nil
	}

	user, createErr := s.userProfile.CreateUserProfile(ctx, phone)
	if createErr == nil {
		return user.ID, nil
	}

	phoneUser, err = s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return bson.NilObjectID, errcode.DatabaseError.WithError(err)
	}
	if phoneUser == nil {
		return bson.NilObjectID, createErr
	}
	if phoneUser.ID.IsZero() {
		return bson.NilObjectID, errcode.DatabaseError.WithError(fmt.Errorf("phone user id is required"))
	}
	if err := s.userProfile.EnsureUserProfileExt(ctx, phoneUser.ID); err != nil {
		return bson.NilObjectID, err
	}
	return phoneUser.ID, nil
}

func (s *IdentityService) ensureWechatBound(ctx context.Context, userID bson.ObjectID, phone, openID, unionID string, now int64, loginIP string) (bson.ObjectID, error) {
	if userID.IsZero() {
		return bson.NilObjectID, errcode.DatabaseError.WithError(fmt.Errorf("userID is required"))
	}

	existing, err := s.authRepo.FindByOpenID(ctx, authmodel.AuthProviderWechat, openID)
	if err != nil {
		return bson.NilObjectID, errcode.DatabaseError.WithError(err)
	}
	if existing != nil {
		if err := s.ensureAuthPhoneMatches(ctx, existing, phone); err != nil {
			return bson.NilObjectID, err
		}
		return s.touchWechatAuth(ctx, existing, now, loginIP)
	}

	unionAuth, err := s.findWechatAuthByUnionID(ctx, unionID)
	if err != nil {
		return bson.NilObjectID, err
	}
	if unionAuth != nil {
		if err := s.ensureAuthPhoneMatches(ctx, unionAuth, phone); err != nil {
			return bson.NilObjectID, err
		}
		return s.updateWechatAuthOpenID(ctx, unionAuth, phone, openID, unionID, now, loginIP)
	}

	auth := &authmodel.UserAuth{
		UserID:       userID,
		AuthProvider: authmodel.AuthProviderWechat,
		OpenID:       openID,
		UnionID:      unionID,
		LastLoginAt:  now,
		LastLoginIP:  loginIP,
	}
	if err := s.authRepo.Create(ctx, auth); err != nil {
		existing, findErr := s.authRepo.FindByOpenID(ctx, authmodel.AuthProviderWechat, openID)
		if findErr != nil {
			return bson.NilObjectID, errcode.DatabaseError.WithError(findErr)
		}
		if existing != nil {
			if err := s.ensureAuthPhoneMatches(ctx, existing, phone); err != nil {
				return bson.NilObjectID, err
			}
			return s.touchWechatAuth(ctx, existing, now, loginIP)
		}
		unionAuth, unionErr := s.findWechatAuthByUnionID(ctx, unionID)
		if unionErr != nil {
			return bson.NilObjectID, unionErr
		}
		if unionAuth != nil {
			if err := s.ensureAuthPhoneMatches(ctx, unionAuth, phone); err != nil {
				return bson.NilObjectID, err
			}
			return s.updateWechatAuthOpenID(ctx, unionAuth, phone, openID, unionID, now, loginIP)
		}
		return bson.NilObjectID, errcode.DatabaseError.WithError(err)
	}
	return userID, nil
}

func (s *IdentityService) ensureAuthPhoneMatches(ctx context.Context, authRecord *authmodel.UserAuth, phone string) error {
	if authRecord == nil {
		return errcode.DatabaseError.WithError(fmt.Errorf("wechat auth is nil"))
	}
	if authRecord.UserID.IsZero() {
		return errcode.DatabaseError.WithError(fmt.Errorf("wechat auth binding broken"))
	}

	user, err := s.userRepo.FindByID(ctx, authRecord.UserID)
	if err != nil {
		return errcode.DatabaseError.WithError(err)
	}
	if user == nil {
		return errcode.DatabaseError.WithError(fmt.Errorf("wechat auth user not found"))
	}
	if user.Phone != phone {
		return errcode.Forbidden.WithError(fmt.Errorf("wechat auth phone mismatch"))
	}
	return nil
}

func (s *IdentityService) updateWechatAuthOpenID(ctx context.Context, authRecord *authmodel.UserAuth, phone, openID, unionID string, now int64, loginIP string) (bson.ObjectID, error) {
	if authRecord == nil {
		return bson.NilObjectID, errcode.DatabaseError.WithError(fmt.Errorf("wechat auth is nil"))
	}
	if authRecord.UserID.IsZero() {
		return bson.NilObjectID, errcode.DatabaseError.WithError(fmt.Errorf("wechat auth binding broken"))
	}
	if err := s.authRepo.UpdateWechatIdentity(ctx, authRecord.ID, openID, unionID, now, loginIP); err != nil {
		existing, findErr := s.authRepo.FindByOpenID(ctx, authmodel.AuthProviderWechat, openID)
		if findErr != nil {
			return bson.NilObjectID, errcode.DatabaseError.WithError(findErr)
		}
		if existing != nil {
			if err := s.ensureAuthPhoneMatches(ctx, existing, phone); err != nil {
				return bson.NilObjectID, err
			}
			return s.touchWechatAuth(ctx, existing, now, loginIP)
		}
		return bson.NilObjectID, errcode.DatabaseError.WithError(err)
	}
	return authRecord.UserID, nil
}

func (s *IdentityService) touchWechatAuth(ctx context.Context, authRecord *authmodel.UserAuth, now int64, loginIP string) (bson.ObjectID, error) {
	if authRecord == nil {
		return bson.NilObjectID, errcode.DatabaseError.WithError(fmt.Errorf("wechat auth is nil"))
	}
	if authRecord.UserID.IsZero() {
		return bson.NilObjectID, errcode.DatabaseError.WithError(fmt.Errorf("wechat auth binding broken"))
	}
	if err := s.authRepo.TouchLastLogin(ctx, authRecord.ID, now, loginIP); err != nil {
		return bson.NilObjectID, errcode.DatabaseError.WithError(err)
	}
	return authRecord.UserID, nil
}
