package user

import (
	"context"
	"strings"

	"house-manager/internal/model"

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
		return nil, invalidParamf("user_id is required")
	}
	user, err := s.users.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, databasef("find user profile: %w", err)
	}
	if user == nil {
		return nil, notFoundf("user not found")
	}
	profile, err := s.profiles.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, databasef("find user profile ext: %w", err)
	}
	return mapProfile(user, profile), nil
}

func (s *Service) UpdateProfile(ctx context.Context, input UpdateProfileInput) (*Profile, error) {
	if input.UserID.IsZero() {
		return nil, invalidParamf("user_id is required")
	}
	if input.BudgetMin < 0 || input.BudgetMax < 0 {
		return nil, invalidParamf("budget must be non-negative")
	}
	if input.BudgetMax > 0 && input.BudgetMin > input.BudgetMax {
		return nil, invalidParamf("budget_min must be less than or equal to budget_max")
	}
	if rentMode := strings.TrimSpace(input.PreferredRentMode); rentMode != "" && !model.RentMode(rentMode).Valid() {
		return nil, invalidParamf("preferred_rent_mode is invalid")
	}
	userFields := bson.M{
		"nickname": strings.TrimSpace(input.Nickname),
		"avatar":   strings.TrimSpace(input.Avatar),
		"city":     strings.TrimSpace(input.City),
	}
	if err := s.users.UpdateProfileFields(ctx, input.UserID, userFields); err != nil {
		return nil, databasef("update user profile: %w", err)
	}
	profileFields := bson.M{
		"budget_min":          input.BudgetMin,
		"budget_max":          input.BudgetMax,
		"preferred_areas":     compactStrings(input.PreferredAreas),
		"preferred_rent_mode": strings.TrimSpace(input.PreferredRentMode),
		"move_in_plan":        strings.TrimSpace(input.MoveInPlan),
		"remark":              strings.TrimSpace(input.Remark),
	}
	if _, err := s.profiles.UpsertByUserID(ctx, input.UserID, profileFields); err != nil {
		return nil, databasef("update user profile ext: %w", err)
	}
	return s.Profile(ctx, ProfileInput{UserID: input.UserID})
}

func (s *Service) Dashboard(ctx context.Context, input DashboardInput) (*Dashboard, error) {
	if input.UserID.IsZero() {
		return nil, invalidParamf("user_id is required")
	}
	favoriteCount, err := s.favorites.Count(ctx, input.UserID)
	if err != nil {
		return nil, databasef("count favorites: %w", err)
	}
	historyCount, err := s.history.Count(ctx, input.UserID)
	if err != nil {
		return nil, databasef("count history: %w", err)
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

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}
