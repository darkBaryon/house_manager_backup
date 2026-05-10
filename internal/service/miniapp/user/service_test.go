package user

import (
	"context"
	"testing"

	"house-manager/internal/model"
	"house-manager/pkg/errcode"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestUserProfileCombinesUserAndProfileExt(t *testing.T) {
	userID := bson.NewObjectID()
	svc := NewService(
		&fakeUserRepository{user: &model.User{CommonFields: model.CommonFields{ID: userID}, Nickname: "小明", Phone: "13800138000", City: "深圳"}},
		&fakeProfileRepository{profile: &model.UserProfileExt{UserID: userID, BudgetMin: 3000, BudgetMax: 6000, PreferredAreas: []string{"南山"}, PreferredRentMode: "whole"}},
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

func TestUserUpdateProfileValidatesBudget(t *testing.T) {
	userID := bson.NewObjectID()
	svc := NewService(&fakeUserRepository{}, &fakeProfileRepository{}, &fakeCounter{}, &fakeCounter{})

	_, err := svc.UpdateProfile(context.Background(), UpdateProfileInput{UserID: userID, BudgetMin: 7000, BudgetMax: 6000})
	assertUserErrCode(t, err, errcode.InvalidParam.Code)
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
	user   *model.User
	fields bson.M
}

func (f *fakeUserRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	return f.user, nil
}

func (f *fakeUserRepository) UpdateProfileFields(ctx context.Context, userID bson.ObjectID, fields bson.M) error {
	f.fields = fields
	return nil
}

type fakeProfileRepository struct {
	profile *model.UserProfileExt
	fields  bson.M
}

func (f *fakeProfileRepository) FindByUserID(ctx context.Context, userID bson.ObjectID) (*model.UserProfileExt, error) {
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
