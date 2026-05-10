package auth

import (
	"context"
	"fmt"
	"time"

	"house-manager/internal/model"
	repoauth "house-manager/internal/repository/auth"
	"house-manager/pkg/errcode"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// IdentityService 负责微信身份注册/登录与 user_id 绑定。
type IdentityService struct {
	authRepo     *repoauth.UserAuthRepository
	userRepo     *repoauth.UserRepository
	userProfile  *UserProfileService
	wechatClient *WechatClient
}

func NewIdentityService(
	authRepo *repoauth.UserAuthRepository,
	userRepo *repoauth.UserRepository,
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

	openID, _, err := s.wechatClient.ExchangeLoginCode(ctx, code)
	if err != nil {
		return bson.NilObjectID, errcode.InternalError.WithError(err)
	}

	authRecord, err := s.authRepo.FindByOpenID(ctx, model.AuthProviderWechat, openID)
	if err != nil {
		return bson.NilObjectID, errcode.DatabaseError.WithError(err)
	}
	if authRecord == nil {
		return bson.NilObjectID, errcode.Unauthorized.WithError(fmt.Errorf("wechat account is not registered"))
	}

	if err := s.authRepo.TouchLastLogin(ctx, authRecord.ID, time.Now().Unix(), loginIP); err != nil {
		return bson.NilObjectID, errcode.DatabaseError.WithError(err)
	}
	return authRecord.UserID, nil
}

func (s *IdentityService) WechatRegister(ctx context.Context, code, phoneCode, loginIP string) (bson.ObjectID, error) {
	if code == "" {
		return bson.NilObjectID, errcode.InvalidParam.WithError(fmt.Errorf("code is required"))
	}
	if phoneCode == "" {
		return bson.NilObjectID, errcode.InvalidParam.WithError(fmt.Errorf("phone code is required"))
	}

	openID, unionID, err := s.wechatClient.ExchangeLoginCode(ctx, code)
	if err != nil {
		return bson.NilObjectID, errcode.InternalError.WithError(err)
	}
	phone, err := s.wechatClient.ExchangePhoneCode(ctx, phoneCode)
	if err != nil {
		return bson.NilObjectID, errcode.InternalError.WithError(err)
	}

	now := time.Now().Unix()

	wechatAuth, err := s.authRepo.FindByOpenID(ctx, model.AuthProviderWechat, openID)
	if err != nil {
		return bson.NilObjectID, errcode.DatabaseError.WithError(err)
	}
	if wechatAuth != nil {
		if err := s.ensureAuthPhoneMatches(ctx, wechatAuth, phone); err != nil {
			return bson.NilObjectID, err
		}
		return s.touchWechatAuth(ctx, wechatAuth, now, loginIP)
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

	userID, err := s.findOrCreateUserByPhone(ctx, phone)
	if err != nil {
		return bson.NilObjectID, err
	}

	return s.ensureWechatBound(ctx, userID, phone, openID, unionID, now, loginIP)
}

func (s *IdentityService) findWechatAuthByUnionID(ctx context.Context, unionID string) (*model.UserAuth, error) {
	if unionID == "" {
		return nil, nil
	}
	authRecord, err := s.authRepo.FindByUnionID(ctx, model.AuthProviderWechat, unionID)
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

	existing, err := s.authRepo.FindByOpenID(ctx, model.AuthProviderWechat, openID)
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

	auth := &model.UserAuth{
		UserID:       userID,
		AuthProvider: model.AuthProviderWechat,
		OpenID:       openID,
		UnionID:      unionID,
		LastLoginAt:  now,
		LastLoginIP:  loginIP,
	}
	if err := s.authRepo.Create(ctx, auth); err != nil {
		existing, findErr := s.authRepo.FindByOpenID(ctx, model.AuthProviderWechat, openID)
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

func (s *IdentityService) ensureAuthPhoneMatches(ctx context.Context, authRecord *model.UserAuth, phone string) error {
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

func (s *IdentityService) updateWechatAuthOpenID(ctx context.Context, authRecord *model.UserAuth, phone, openID, unionID string, now int64, loginIP string) (bson.ObjectID, error) {
	if authRecord == nil {
		return bson.NilObjectID, errcode.DatabaseError.WithError(fmt.Errorf("wechat auth is nil"))
	}
	if authRecord.UserID.IsZero() {
		return bson.NilObjectID, errcode.DatabaseError.WithError(fmt.Errorf("wechat auth binding broken"))
	}
	if err := s.authRepo.UpdateWechatIdentity(ctx, authRecord.ID, openID, unionID, now, loginIP); err != nil {
		existing, findErr := s.authRepo.FindByOpenID(ctx, model.AuthProviderWechat, openID)
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

func (s *IdentityService) touchWechatAuth(ctx context.Context, authRecord *model.UserAuth, now int64, loginIP string) (bson.ObjectID, error) {
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
