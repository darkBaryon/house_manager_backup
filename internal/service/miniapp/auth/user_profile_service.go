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
	slog.InfoContext(ctx, "miniapp.auth.user_profile.create.start", miniapplog.Attrs(ctx,
		"phone", miniapplog.MaskPhone(phone),
	)...)

	now := time.Now().Unix()
	user := &authmodel.User{
		Phone:         phone,
		SourceChannel: authmodel.SourceChannelMini,
		LastActiveAt:  now,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.user_profile.create.failed", miniapplog.Attrs(ctx,
			"phone", miniapplog.MaskPhone(phone),
			"step", "create_user",
			"error", err,
		)...)
		return nil, errcode.DatabaseError.WithError(err)
	}

	profile := &authmodel.UserProfileExt{UserID: user.ID}
	if err := s.profileExtRepo.Create(ctx, profile); err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.user_profile.create.failed", miniapplog.Attrs(ctx,
			"phone", miniapplog.MaskPhone(phone),
			"user_id", user.ID.Hex(),
			"step", "create_profile_ext",
			"error", err,
		)...)
		return nil, errcode.DatabaseError.WithError(err)
	}
	slog.InfoContext(ctx, "miniapp.auth.user_profile.create.success", miniapplog.Attrs(ctx,
		"phone", miniapplog.MaskPhone(phone),
		"user_id", user.ID.Hex(),
	)...)
	return user, nil
}

func (s *UserProfileService) EnsureUserProfileExt(ctx context.Context, userID bson.ObjectID) error {
	if userID.IsZero() {
		return errcode.InvalidParam.WithError(fmt.Errorf("userID is required"))
	}
	slog.InfoContext(ctx, "miniapp.auth.user_profile.ensure_ext.start", miniapplog.Attrs(ctx,
		"user_id", userID.Hex(),
	)...)
	if _, err := s.profileExtRepo.UpsertByUserID(ctx, userID, bson.M{}); err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.user_profile.ensure_ext.failed", miniapplog.Attrs(ctx,
			"user_id", userID.Hex(),
			"error", err,
		)...)
		return errcode.DatabaseError.WithError(err)
	}
	slog.InfoContext(ctx, "miniapp.auth.user_profile.ensure_ext.success", miniapplog.Attrs(ctx,
		"user_id", userID.Hex(),
	)...)
	return nil
}

func (s *UserProfileService) TouchUserLastActive(ctx context.Context, userID bson.ObjectID) error {
	if userID.IsZero() {
		return errcode.InvalidParam.WithError(fmt.Errorf("userID is required"))
	}
	if err := s.userRepo.TouchLastActive(ctx, userID, time.Now().Unix()); err != nil {
		slog.ErrorContext(ctx, "miniapp.auth.user_profile.touch_last_active.failed", miniapplog.Attrs(ctx,
			"user_id", userID.Hex(),
			"error", err,
		)...)
		return errcode.DatabaseError.WithError(err)
	}
	return nil
}
