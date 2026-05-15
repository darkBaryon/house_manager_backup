package user

import (
	"context"
	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/pkg/errcode"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestUserProfileCombinesUserAndProfileExt(t *testing.T) {
	userID := bson.NewObjectID()
	svc := NewService(
		&fakeUserRepository{user: &authmodel.User{CommonFields: commonmodel.CommonFields{ID: userID}, Nickname: "小明", Phone: "13800138000", City: "深圳"}},
		&fakeProfileRepository{profile: &authmodel.UserProfileExt{UserID: userID, BudgetMin: 3000, BudgetMax: 6000, PreferredAreas: []string{"南山"}, PreferredRentMode: "whole"}},
		&fakeCounter{},
		&fakeCounter{},
	)

	profile, err := svc.Profile(context.Background(), ProfileInput{UserID: userID})
	if err != nil {
		t.Fatalf("Profile returned error: %v", err)
	}
	if profile.UserID != userID.Hex() || profile.Phone != "13800138000" || profile.BudgetMax != 6000 || len(profile.PreferredAreas) != 1 {
		t.Fatalf("unexpected profile: %#v", profile)
	}
}

func TestUserProfileAllowsMissingExt(t *testing.T) {
	userID := bson.NewObjectID()
	svc := NewService(
		&fakeUserRepository{user: &authmodel.User{CommonFields: commonmodel.CommonFields{ID: userID}, Nickname: "小明"}},
		&fakeProfileRepository{},
		&fakeCounter{},
		&fakeCounter{},
	)

	profile, err := svc.Profile(context.Background(), ProfileInput{UserID: userID})
	if err != nil {
		t.Fatalf("Profile returned error: %v", err)
	}
	if profile.UserID != userID.Hex() || len(profile.PreferredAreas) != 0 || profile.BudgetMin != 0 || profile.BudgetMax != 0 {
		t.Fatalf("unexpected empty ext profile: %#v", profile)
	}
}

func TestUserUpdateProfileValidatesBudget(t *testing.T) {
	userID := bson.NewObjectID()
	svc := NewService(&fakeUserRepository{}, &fakeProfileRepository{}, &fakeCounter{}, &fakeCounter{})

	_, err := svc.UpdateProfile(context.Background(), UpdateProfileInput{UserID: userID, BudgetMin: intPtr(7000), BudgetMax: intPtr(6000)})
	assertUserErrCode(t, err, errcode.InvalidParam.Code)
}

func TestUserUpdateProfileOnlyWritesProvidedFields(t *testing.T) {
	userID := bson.NewObjectID()
	users := &fakeUserRepository{user: &authmodel.User{CommonFields: commonmodel.CommonFields{ID: userID}, Nickname: "旧昵称", City: "深圳"}}
	profiles := &fakeProfileRepository{profile: &authmodel.UserProfileExt{UserID: userID, BudgetMin: 1000, BudgetMax: 5000, Remark: "旧备注"}}
	svc := NewService(users, profiles, &fakeCounter{}, &fakeCounter{})

	profile, err := svc.UpdateProfile(context.Background(), UpdateProfileInput{
		UserID:    userID,
		Nickname:  stringPtr(" 新昵称 "),
		BudgetMin: intPtr(2000),
	})
	if err != nil {
		t.Fatalf("UpdateProfile returned error: %v", err)
	}
	if users.fields["nickname"] != "新昵称" {
		t.Fatalf("unexpected user fields: %#v", users.fields)
	}
	if _, ok := users.fields["avatar"]; ok {
		t.Fatalf("unprovided avatar must not be written: %#v", users.fields)
	}
	if profiles.fields["budget_min"] != 2000 {
		t.Fatalf("unexpected profile fields: %#v", profiles.fields)
	}
	if _, ok := profiles.fields["remark"]; ok {
		t.Fatalf("unprovided remark must not be written: %#v", profiles.fields)
	}
	if profile == nil {
		t.Fatalf("expected profile response")
	}
}

func TestUserDashboardCountsFavoriteAndHistory(t *testing.T) {
	userID := bson.NewObjectID()
	svc := NewService(&fakeUserRepository{}, &fakeProfileRepository{}, &fakeCounter{count: 3}, &fakeCounter{count: 12})

	dashboard, err := svc.Dashboard(context.Background(), DashboardInput{UserID: userID})
	if err != nil {
		t.Fatalf("Dashboard returned error: %v", err)
	}
	if dashboard.FavoriteCount != 3 || dashboard.HistoryCount != 12 || dashboard.PlanCount != 0 || dashboard.UnreadNotificationCount != 0 {
		t.Fatalf("unexpected dashboard: %#v", dashboard)
	}
}

func assertUserErrCode(t *testing.T, err error, code int) {
	t.Helper()
	e := errcode.FromError(err)
	if e == nil || e.Code != code {
		t.Fatalf("expected error code %d, got %v from %v", code, e, err)
	}
}

type fakeUserRepository struct {
	user   *authmodel.User
	fields bson.M
}

func (f *fakeUserRepository) FindByID(ctx context.Context, id bson.ObjectID) (*authmodel.User, error) {
	return f.user, nil
}

func (f *fakeUserRepository) UpdateProfileFields(ctx context.Context, userID bson.ObjectID, fields bson.M) error {
	f.fields = fields
	return nil
}

type fakeProfileRepository struct {
	profile *authmodel.UserProfileExt
	fields  bson.M
}

func (f *fakeProfileRepository) FindByUserID(ctx context.Context, userID bson.ObjectID) (*authmodel.UserProfileExt, error) {
	return f.profile, nil
}

func (f *fakeProfileRepository) UpsertByUserID(ctx context.Context, userID bson.ObjectID, fields bson.M) (bool, error) {
	f.fields = fields
	return true, nil
}

type fakeCounter struct {
	count int64
}

func (f *fakeCounter) Count(ctx context.Context, userID bson.ObjectID) (int64, error) {
	return f.count, nil
}

func stringPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
}
