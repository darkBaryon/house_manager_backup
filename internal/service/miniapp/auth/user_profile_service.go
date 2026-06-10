package auth

import (
	"context"
	"fmt"
	authmodel "house-manager/internal/model/auth"
	miniappauthrepo "house-manager/internal/repository/miniapp_auth"
	"house-manager/pkg/applog"
	"house-manager/pkg/errcode"
	"log/slog"
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
	slog.InfoContext(ctx, "miniapp.auth.user_profile.create.start", "phone", applog.Phone(phone))

	now := time.Now().Unix()
	user := &authmodel.User{
		Phone:         phone,
		SourceChannel: authmodel.SourceChannelMini,
		LastActiveAt:  now,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.user_profile.create.failed", "phone", applog.Phone(phone),
			"step", "create_user",
			"error", err,
		)
		return nil, errcode.DatabaseError.WithError(err)
	}

	profile := &authmodel.UserProfileExt{UserID: user.ID}
	if err := s.profileExtRepo.Create(ctx, profile); err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.user_profile.create.failed", "phone", applog.Phone(phone),
			"user_id", user.ID.Hex(),
			"step", "create_profile_ext",
			"error", err,
		)
		return nil, errcode.DatabaseError.WithError(err)
	}
	slog.InfoContext(ctx, "miniapp.auth.user_profile.create.success", "phone", applog.Phone(phone),
		"user_id", user.ID.Hex(),
	)
	return user, nil
}

func (s *UserProfileService) EnsureUserProfileExt(ctx context.Context, userID bson.ObjectID) error {
	if userID.IsZero() {
		return errcode.InvalidParam.WithError(fmt.Errorf("userID is required"))
	}
	slog.InfoContext(ctx, "miniapp.auth.user_profile.ensure_ext.start", "user_id", userID.Hex())
	if _, err := s.profileExtRepo.UpsertByUserID(ctx, userID, bson.M{}); err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.user_profile.ensure_ext.failed", "user_id", userID.Hex(),
			"error", err,
		)
		return errcode.DatabaseError.WithError(err)
	}
	slog.InfoContext(ctx, "miniapp.auth.user_profile.ensure_ext.success", "user_id", userID.Hex())
	return nil
}

func (s *UserProfileService) TouchUserLastActive(ctx context.Context, userID bson.ObjectID) error {
	if userID.IsZero() {
		return errcode.InvalidParam.WithError(fmt.Errorf("userID is required"))
	}
	if err := s.userRepo.TouchLastActive(ctx, userID, time.Now().Unix()); err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.user_profile.touch_last_active.failed", "user_id", userID.Hex(),
			"error", err,
		)
		return errcode.DatabaseError.WithError(err)
	}
	return nil
}
