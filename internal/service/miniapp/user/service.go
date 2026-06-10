package user

import (
	"context"
	authmodel "house-manager/internal/model/auth"
	hmdmodel "house-manager/internal/model/hmd"
	"house-manager/pkg/errcode"
	"log/slog"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	users     userRepository
	profiles  profileRepository
	favorites counter
	history   counter
}

type userRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*authmodel.User, error)
	UpdateProfileFields(ctx context.Context, userID bson.ObjectID, fields bson.M) error
}

type profileRepository interface {
	FindByUserID(ctx context.Context, userID bson.ObjectID) (*authmodel.UserProfileExt, error)
	UpsertByUserID(ctx context.Context, userID bson.ObjectID, fields bson.M) (bool, error)
}

type counter interface {
	Count(ctx context.Context, userID bson.ObjectID) (int64, error)
}

func NewService(users userRepository, profiles profileRepository, favorites counter, history counter) *Service {
	return &Service{users: users, profiles: profiles, favorites: favorites, history: history}
}

func (s *Service) Profile(ctx context.Context, input ProfileInput) (*Profile, error) {
	slog.InfoContext(ctx, "miniapp.user.profile.start", "user_id", input.UserID.Hex())
	if input.UserID.IsZero() {
		err := errcode.InvalidParam.WithErrorf("user_id is required")
		slog.WarnContext(ctx, "miniapp.user.profile.failed", "error", err)
		return nil, err
	}
	user, err := s.users.FindByID(ctx, input.UserID)
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("find user profile: %w", err)
		slog.ErrorContext(ctx, "miniapp.user.profile.failed", "user_id", input.UserID.Hex(),
			"step", "find_user",
			"error", wrapped,
		)
		return nil, wrapped
	}
	if user == nil {
		err := errcode.NotFound.WithErrorf("user not found")
		slog.WarnContext(ctx, "miniapp.user.profile.failed", "user_id", input.UserID.Hex(),
			"step", "find_user",
			"error", err,
		)
		return nil, err
	}
	profile, err := s.profiles.FindByUserID(ctx, input.UserID)
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("find user profile ext: %w", err)
		slog.ErrorContext(ctx, "miniapp.user.profile.failed", "user_id", input.UserID.Hex(),
			"step", "find_profile_ext",
			"error", wrapped,
		)
		return nil, wrapped
	}
	result := mapProfile(user, profile)
	slog.InfoContext(ctx, "miniapp.user.profile.success", "user_id", input.UserID.Hex(),
		"city", result.City,
	)
	return result, nil
}

func (s *Service) UpdateProfile(ctx context.Context, input UpdateProfileInput) (*Profile, error) {
	slog.InfoContext(ctx, "miniapp.user.update_profile.start", "user_id", input.UserID.Hex(),
		"has_nickname", input.Nickname != nil,
		"has_city", input.City != nil,
		"has_budget_min", input.BudgetMin != nil,
		"has_budget_max", input.BudgetMax != nil,
		"has_preferred_areas", input.PreferredAreas != nil,
	)
	if input.UserID.IsZero() {
		err := errcode.InvalidParam.WithErrorf("user_id is required")
		slog.WarnContext(ctx, "miniapp.user.update_profile.failed", "error", err)
		return nil, err
	}
	if (input.BudgetMin != nil && *input.BudgetMin < 0) || (input.BudgetMax != nil && *input.BudgetMax < 0) {
		err := errcode.InvalidParam.WithErrorf("budget must be non-negative")
		slog.WarnContext(ctx, "miniapp.user.update_profile.failed", "user_id", input.UserID.Hex(),
			"error", err,
		)
		return nil, err
	}
	if rentMode := trimStringPtr(input.PreferredRentMode); rentMode != "" && !hmdmodel.RentMode(rentMode).Valid() {
		err := errcode.InvalidParam.WithErrorf("preferred_rent_mode is invalid")
		slog.WarnContext(ctx, "miniapp.user.update_profile.failed", "user_id", input.UserID.Hex(),
			"preferred_rent_mode", rentMode,
			"error", err,
		)
		return nil, err
	}

	profile, err := s.profiles.FindByUserID(ctx, input.UserID)
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("find user profile ext: %w", err)
		slog.ErrorContext(ctx, "miniapp.user.update_profile.failed", "user_id", input.UserID.Hex(),
			"step", "find_profile_ext",
			"error", wrapped,
		)
		return nil, wrapped
	}
	effectiveBudgetMin, effectiveBudgetMax := effectiveBudgetRange(profile, input)
	if effectiveBudgetMax > 0 && effectiveBudgetMin > effectiveBudgetMax {
		err := errcode.InvalidParam.WithErrorf("budget_min must be less than or equal to budget_max")
		slog.WarnContext(ctx, "miniapp.user.update_profile.failed", "user_id", input.UserID.Hex(),
			"error", err,
		)
		return nil, err
	}

	userFields := bson.M{}
	setTrimmedString(userFields, "nickname", input.Nickname)
	setTrimmedString(userFields, "avatar", input.Avatar)
	setTrimmedString(userFields, "city", input.City)
	if len(userFields) > 0 {
		if err := s.users.UpdateProfileFields(ctx, input.UserID, userFields); err != nil {
			wrapped := errcode.DatabaseError.WithErrorf("update user profile: %w", err)
			slog.ErrorContext(ctx, "miniapp.user.update_profile.failed", "user_id", input.UserID.Hex(),
				"step", "update_user_profile",
				"error", wrapped,
			)
			return nil, wrapped
		}
	}

	profileFields := bson.M{}
	if input.BudgetMin != nil {
		profileFields["budget_min"] = *input.BudgetMin
	}
	if input.BudgetMax != nil {
		profileFields["budget_max"] = *input.BudgetMax
	}
	if input.PreferredAreas != nil {
		profileFields["preferred_areas"] = compactStrings(*input.PreferredAreas)
	}
	setTrimmedString(profileFields, "preferred_rent_mode", input.PreferredRentMode)
	setTrimmedString(profileFields, "move_in_plan", input.MoveInPlan)
	setTrimmedString(profileFields, "remark", input.Remark)
	if len(profileFields) > 0 {
		if _, err := s.profiles.UpsertByUserID(ctx, input.UserID, profileFields); err != nil {
			wrapped := errcode.DatabaseError.WithErrorf("update user profile ext: %w", err)
			slog.ErrorContext(ctx, "miniapp.user.update_profile.failed", "user_id", input.UserID.Hex(),
				"step", "upsert_profile_ext",
				"error", wrapped,
			)
			return nil, wrapped
		}
	}
	result, err := s.Profile(ctx, ProfileInput{UserID: input.UserID})
	if err != nil {
		slog.WarnContext(ctx, "miniapp.user.update_profile.failed", "user_id", input.UserID.Hex(),
			"step", "reload_profile",
			"error", err,
		)
		return nil, err
	}
	slog.InfoContext(ctx, "miniapp.user.update_profile.success", "user_id", input.UserID.Hex(),
		"city", result.City,
	)
	return result, nil
}

func (s *Service) Dashboard(ctx context.Context, input DashboardInput) (*Dashboard, error) {
	slog.InfoContext(ctx, "miniapp.user.dashboard.start", "user_id", input.UserID.Hex())
	if input.UserID.IsZero() {
		err := errcode.InvalidParam.WithErrorf("user_id is required")
		slog.WarnContext(ctx, "miniapp.user.dashboard.failed", "error", err)
		return nil, err
	}
	favoriteCount, err := s.favorites.Count(ctx, input.UserID)
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("count favorites: %w", err)
		slog.WarnContext(ctx, "miniapp.user.dashboard.failed", "user_id", input.UserID.Hex(),
			"step", "count_favorites",
			"error", wrapped,
		)
		return nil, wrapped
	}
	historyCount, err := s.history.Count(ctx, input.UserID)
	if err != nil {
		wrapped := errcode.DatabaseError.WithErrorf("count history: %w", err)
		slog.WarnContext(ctx, "miniapp.user.dashboard.failed", "user_id", input.UserID.Hex(),
			"step", "count_history",
			"error", wrapped,
		)
		return nil, wrapped
	}
	result := &Dashboard{
		FavoriteCount:           favoriteCount,
		HistoryCount:            historyCount,
		PlanCount:               0,
		UnreadNotificationCount: 0,
	}
	slog.InfoContext(ctx, "miniapp.user.dashboard.success", "user_id", input.UserID.Hex(),
		"favorite_count", favoriteCount,
		"history_count", historyCount,
	)
	return result, nil
}

func mapProfile(user *authmodel.User, ext *authmodel.UserProfileExt) *Profile {
	out := &Profile{
		UserID:         user.ID.Hex(),
		Nickname:       user.Nickname,
		Avatar:         user.Avatar,
		Phone:          user.Phone,
		City:           user.City,
		PreferredAreas: []string{},
	}
	if ext != nil {
		out.BudgetMin = ext.BudgetMin
		out.BudgetMax = ext.BudgetMax
		out.PreferredAreas = cloneStrings(ext.PreferredAreas)
		out.PreferredRentMode = ext.PreferredRentMode
		out.MoveInPlan = ext.MoveInPlan
		out.Remark = ext.Remark
	}
	return out
}

func compactStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func trimStringPtr(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func setTrimmedString(fields bson.M, key string, value *string) {
	if value != nil {
		fields[key] = strings.TrimSpace(*value)
	}
}

func effectiveBudgetRange(profile *authmodel.UserProfileExt, input UpdateProfileInput) (int, int) {
	minValue, maxValue := 0, 0
	if profile != nil {
		minValue = profile.BudgetMin
		maxValue = profile.BudgetMax
	}
	if input.BudgetMin != nil {
		minValue = *input.BudgetMin
	}
	if input.BudgetMax != nil {
		maxValue = *input.BudgetMax
	}
	return minValue, maxValue
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}
