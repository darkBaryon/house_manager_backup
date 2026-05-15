package auth

import (
	"context"
	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestLoginRejectsPhoneOnlyOutsideLocal(t *testing.T) {
	svc := NewService(nil, nil, "prod")

	_, err := svc.Login(context.Background(), LoginInput{Phone: "13800000000"})

	got := errcode.FromError(err)
	if got == nil || got.Code != errcode.Forbidden.Code {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestLoginRequiresPhone(t *testing.T) {
	svc := NewService(nil, nil, "local")

	_, err := svc.Login(context.Background(), LoginInput{})

	got := errcode.FromError(err)
	if got == nil || got.Code != errcode.InvalidParam.Code {
		t.Fatalf("expected invalid param, got %v", err)
	}
}

func TestSessionRejectsNonPublishPrincipal(t *testing.T) {
	svc := NewService(nil, nil, "local")

	_, err := svc.Session(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeUser,
		PrincipalID:   "user-id",
		Terminal:      session.TerminalMiniapp,
	})

	got := errcode.FromError(err)
	if got == nil || got.Code != errcode.Unauthorized.Code {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestLoginBuildsUserPrincipal(t *testing.T) {
	userID := bson.NewObjectID()
	store := session.NewStore(newFakeCache(), 0)
	userRepo := &fakeOwnerUserRepository{
		activeByPhone: &authmodel.User{
			CommonFields: commonmodel.CommonFields{ID: userID, Status: commonmodel.StatusActive},
			Phone:        "13800000000",
			Nickname:     "房东A",
		},
	}
	svc := newService(userRepo, store, "local")

	result, err := svc.Login(context.Background(), LoginInput{Phone: "13800000000", LoginIP: "127.0.0.1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected token")
	}
	if result.Principal.PrincipalType != session.PrincipalTypeUser || result.Principal.Terminal != session.TerminalPublish {
		t.Fatalf("unexpected principal: %#v", result.Principal)
	}
	if result.Principal.PrincipalID != userID.Hex() || result.Principal.Phone != "13800000000" {
		t.Fatalf("unexpected principal identity: %#v", result.Principal)
	}
	if len(result.Principal.RoleCodes) != 0 || len(result.Principal.PermissionCodes) != 0 {
		t.Fatalf("expected no role or permission codes, got %#v", result.Principal)
	}
	if result.Subject.Name != "房东A" {
		t.Fatalf("unexpected subject: %#v", result.Subject)
	}
}

type fakeOwnerUserRepository struct {
	activeByPhone *authmodel.User
	byID          *authmodel.User
}

func (f *fakeOwnerUserRepository) FindActiveByPhone(ctx context.Context, phone string) (*authmodel.User, error) {
	return f.activeByPhone, nil
}

func (f *fakeOwnerUserRepository) FindByID(ctx context.Context, id bson.ObjectID) (*authmodel.User, error) {
	if f.byID != nil {
		return f.byID, nil
	}
	return f.activeByPhone, nil
}

type fakeCache struct {
	values map[string]string
}

func newFakeCache() *fakeCache {
	return &fakeCache{values: map[string]string{}}
}

func (f *fakeCache) Get(ctx context.Context, key string) (string, error) {
	return f.values[key], nil
}

func (f *fakeCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	f.values[key] = value
	return nil
}

func (f *fakeCache) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(f.values, key)
	}
	return nil
}
