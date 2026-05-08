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

// UserProfileService 负责用户主档初始化与活跃时间维护。
type UserProfileService struct {
	userRepo       *repoauth.UserRepository
	profileExtRepo *repoauth.UserProfileExtRepository
}

func NewUserProfileService(
	userRepo *repoauth.UserRepository,
	profileExtRepo *repoauth.UserProfileExtRepository,
) *UserProfileService {
	return &UserProfileService{userRepo: userRepo, profileExtRepo: profileExtRepo}
}

func (s *UserProfileService) CreateUserProfile(ctx context.Context, phone string) (*model.User, error) {
	if phone == "" {
		return nil, errcode.InvalidParam.WithError(fmt.Errorf("phone is required"))
	}

	now := time.Now().Unix()
	user := &model.User{
		Phone:         phone,
		SourceChannel: model.SourceChannelMini,
		LastActiveAt:  now,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errcode.DatabaseError.WithError(err)
	}

	profile := &model.UserProfileExt{UserID: user.ID}
	if err := s.profileExtRepo.Create(ctx, profile); err != nil {
		return nil, errcode.DatabaseError.WithError(err)
	}
	return user, nil
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
