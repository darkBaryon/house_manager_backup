package auth

import (
	"context"
	"fmt"
	"time"

	"house-manager/internal/model"
	"house-manager/internal/repository"
	dbmongo "house-manager/pkg/database/mongo"
	"house-manager/pkg/errcode"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// IdentityService 负责微信身份注册/登录与 user_id 绑定。
type IdentityService struct {
	mongoClient  *dbmongo.Client
	authRepo     *repository.UserAuthRepository
	userRepo     *repository.UserRepository
	userProfile  *UserProfileService
	wechatClient *WechatClient
}

func NewIdentityService(
	mongoClient *dbmongo.Client,
	authRepo *repository.UserAuthRepository,
	userRepo *repository.UserRepository,
	userProfile *UserProfileService,
	wechatClient *WechatClient,
) *IdentityService {
	return &IdentityService{
		mongoClient:  mongoClient,
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

	wechatAuth, err := s.authRepo.FindByOpenID(ctx, model.AuthProviderWechat, openID)
	if err != nil {
		return bson.NilObjectID, errcode.DatabaseError.WithError(err)
	}
	phoneUser, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return bson.NilObjectID, errcode.DatabaseError.WithError(err)
	}

	if wechatAuth != nil && wechatAuth.UserID.IsZero() {
		return bson.NilObjectID, errcode.DatabaseError.WithError(fmt.Errorf("wechat auth binding broken"))
	}

	var userID bson.ObjectID
	switch {
	case wechatAuth != nil:
		userID = wechatAuth.UserID
	case phoneUser != nil:
		userID = phoneUser.ID
	default:
		userID = bson.NilObjectID
	}

	now := time.Now().Unix()
	if userID.IsZero() {
		user, createErr := s.userProfile.CreateUserProfile(ctx, phone)
		if createErr != nil {
			return bson.NilObjectID, createErr
		}
		userID = user.ID
	}

	if err := s.ensureWechatBound(ctx, userID, openID, unionID, now, loginIP); err != nil {
		return bson.NilObjectID, err
	}

	return userID, nil
}

func (s *IdentityService) ensureWechatBound(ctx context.Context, userID bson.ObjectID, openID, unionID string, now int64, loginIP string) error {
	if userID.IsZero() {
		return errcode.DatabaseError.WithError(fmt.Errorf("userID is required"))
	}

	existing, err := s.authRepo.FindByOpenID(ctx, model.AuthProviderWechat, openID)
	if err != nil {
		return errcode.DatabaseError.WithError(err)
	}
	if existing != nil {
		if existing.UserID != userID {
			return errcode.Forbidden.WithError(fmt.Errorf("wechat account already bound to another user"))
		}
		return s.authRepo.TouchLastLogin(ctx, existing.ID, now, loginIP)
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
		return errcode.DatabaseError.WithError(err)
	}
	return nil
}
