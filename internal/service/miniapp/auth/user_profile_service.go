package auth

import (
	"context"
	"fmt"
	authmodel "house-manager/internal/model/auth"
	miniappauthrepo "house-manager/internal/repository/miniapp_auth"
	"house-manager/pkg/errcode"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// UserProfileService 负责用户主档初始化与活跃时间维护。
type UserProfileService struct {
	userRepo       *miniappauthrepo.UserRepository
	profileExtRepo *miniappauthrepo.UserProfileExtRepository
}

func NewUserProfileService(
	userRepo *miniappauthrepo.UserRepository,
	profileExtRepo *miniappauthrepo.UserProfileExtRepository,
) *UserProfileService {
	return &UserProfileService{userRepo: userRepo, profileExtRepo: profileExtRepo}
}

func (s *UserProfileService) CreateUserProfile(ctx context.Context, phone string) (*authmodel.User, error) {
	if phone == "" {
		return nil, errcode.InvalidParam.WithError(fmt.Errorf("phone is required"))
	}

	now := time.Now().Unix()
	user := &authmodel.User{
		Phone:         phone,
		SourceChannel: authmodel.SourceChannelMini,
		LastActiveAt:  now,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errcode.DatabaseError.WithError(err)
	}

	profile := &authmodel.UserProfileExt{UserID: user.ID}
	if err := s.profileExtRepo.Create(ctx, profile); err != nil {
		return nil, errcode.DatabaseError.WithError(err)
	}
	return user, nil
}

func (s *UserProfileService) EnsureUserProfileExt(ctx context.Context, userID bson.ObjectID) error {
	if userID.IsZero() {
		return errcode.InvalidParam.WithError(fmt.Errorf("userID is required"))
	}
	if _, err := s.profileExtRepo.UpsertByUserID(ctx, userID, bson.M{}); err != nil {
		return errcode.DatabaseError.WithError(err)
	}
	return nil
}

func (s *UserProfileService) TouchUserLastActive(ctx context.Context, userID bson.ObjectID) error {
	if userID.IsZero() {
		return errcode.InvalidParam.WithError(fmt.Errorf("userID is required"))
	}
	if err := s.userRepo.TouchLastActive(ctx, userID, time.Now().Unix()); err != nil {
		return errcode.DatabaseError.WithError(err)
	}
	return nil
}
