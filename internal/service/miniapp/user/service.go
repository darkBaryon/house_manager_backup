package user

import (
	"context"
	"strings"

	"house-manager/internal/model"
	"house-manager/pkg/errcode"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	users     userRepository
	profiles  profileRepository
	favorites counter
	history   counter
}

type userRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*model.User, error)
	UpdateProfileFields(ctx context.Context, userID bson.ObjectID, fields bson.M) error
}

type profileRepository interface {
	FindByUserID(ctx context.Context, userID bson.ObjectID) (*model.UserProfileExt, error)
	UpsertByUserID(ctx context.Context, userID bson.ObjectID, fields bson.M) (bool, error)
}

type counter interface {
	Count(ctx context.Context, userID bson.ObjectID) (int64, error)
}

func NewService(users userRepository, profiles profileRepository, favorites counter, history counter) *Service {
	return &Service{users: users, profiles: profiles, favorites: favorites, history: history}
}

func (s *Service) Profile(ctx context.Context, input ProfileInput) (*Profile, error) {
	if input.UserID.IsZero() {
		return nil, errcode.InvalidParam.WithErrorf("user_id is required")
	}
	user, err := s.users.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, errcode.DatabaseError.WithErrorf("find user profile: %w", err)
	}
	if user == nil {
		return nil, errcode.NotFound.WithErrorf("user not found")
	}
	profile, err := s.profiles.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, errcode.DatabaseError.WithErrorf("find user profile ext: %w", err)
	}
	return mapProfile(user, profile), nil
}

func (s *Service) UpdateProfile(ctx context.Context, input UpdateProfileInput) (*Profile, error) {
	if input.UserID.IsZero() {
		return nil, errcode.InvalidParam.WithErrorf("user_id is required")
	}
	if (input.BudgetMin != nil && *input.BudgetMin < 0) || (input.BudgetMax != nil && *input.BudgetMax < 0) {
		return nil, errcode.InvalidParam.WithErrorf("budget must be non-negative")
	}
	if rentMode := trimStringPtr(input.PreferredRentMode); rentMode != "" && !model.RentMode(rentMode).Valid() {
		return nil, errcode.InvalidParam.WithErrorf("preferred_rent_mode is invalid")
	}

	profile, err := s.profiles.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, errcode.DatabaseError.WithErrorf("find user profile ext: %w", err)
	}
	effectiveBudgetMin, effectiveBudgetMax := effectiveBudgetRange(profile, input)
	if effectiveBudgetMax > 0 && effectiveBudgetMin > effectiveBudgetMax {
		return nil, errcode.InvalidParam.WithErrorf("budget_min must be less than or equal to budget_max")
	}

	userFields := bson.M{}
	setTrimmedString(userFields, "nickname", input.Nickname)
	setTrimmedString(userFields, "avatar", input.Avatar)
	setTrimmedString(userFields, "city", input.City)
	if len(userFields) > 0 {
		if err := s.users.UpdateProfileFields(ctx, input.UserID, userFields); err != nil {
			return nil, errcode.DatabaseError.WithErrorf("update user profile: %w", err)
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
			return nil, errcode.DatabaseError.WithErrorf("update user profile ext: %w", err)
		}
	}
	return s.Profile(ctx, ProfileInput{UserID: input.UserID})
}

func (s *Service) Dashboard(ctx context.Context, input DashboardInput) (*Dashboard, error) {
	if input.UserID.IsZero() {
		return nil, errcode.InvalidParam.WithErrorf("user_id is required")
	}
	favoriteCount, err := s.favorites.Count(ctx, input.UserID)
	if err != nil {
		return nil, errcode.DatabaseError.WithErrorf("count favorites: %w", err)
	}
	historyCount, err := s.history.Count(ctx, input.UserID)
	if err != nil {
		return nil, errcode.DatabaseError.WithErrorf("count history: %w", err)
	}
	return &Dashboard{
		FavoriteCount:           favoriteCount,
		HistoryCount:            historyCount,
		PlanCount:               0,
		UnreadNotificationCount: 0,
	}, nil
}

func mapProfile(user *model.User, ext *model.UserProfileExt) *Profile {
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

func effectiveBudgetRange(profile *model.UserProfileExt, input UpdateProfileInput) (int, int) {
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
