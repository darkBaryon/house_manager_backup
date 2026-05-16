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
	"golang.org/x/crypto/bcrypt"
)

func TestLoginRequiresPassword(t *testing.T) {
	svc := newService(nil, nil, nil)

	_, err := svc.Login(context.Background(), LoginInput{Phone: "13800000000"})

	got := errcode.FromError(err)
	if got == nil || got.Code != errcode.InvalidParam.Code {
		t.Fatalf("expected invalid param, got %v", err)
	}
}

func TestLoginRequiresPhone(t *testing.T) {
	svc := newService(nil, nil, nil)

	_, err := svc.Login(context.Background(), LoginInput{})

	got := errcode.FromError(err)
	if got == nil || got.Code != errcode.InvalidParam.Code {
		t.Fatalf("expected invalid param, got %v", err)
	}
}

func TestSessionRejectsNonPublishPrincipal(t *testing.T) {
	svc := newService(nil, nil, nil)

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

func TestLoginBuildsLandlordPrincipal(t *testing.T) {
	landlordID := bson.NewObjectID()
	authID := bson.NewObjectID()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	store := session.NewStore(newFakeCache(), 0)
	landlordRepo := &fakeLandlordRepository{
		activeByPhone: &authmodel.Landlord{
			CommonFields: commonmodel.CommonFields{ID: landlordID, Status: commonmodel.StatusActive},
			Phone:        "13800000000",
		},
		byID: &authmodel.Landlord{
			CommonFields: commonmodel.CommonFields{ID: landlordID, Status: commonmodel.StatusActive},
			Phone:        "13800000000",
		},
	}
	landlordAuthRepo := &fakeLandlordAuthRepository{
		activeByLandlordID: &authmodel.LandlordAuth{
			CommonFields: commonmodel.CommonFields{ID: authID, Status: commonmodel.StatusActive},
			LandlordID:   landlordID,
			AuthType:     authmodel.PasswordAuthTypePassword,
			PasswordHash: string(hash),
		},
	}
	svc := newService(landlordRepo, landlordAuthRepo, store)

	result, err := svc.Login(context.Background(), LoginInput{Phone: "13800000000", Password: "secret123", LoginIP: "127.0.0.1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected token")
	}
	if result.Principal.PrincipalType != session.PrincipalTypeLandlord || result.Principal.Terminal != session.TerminalPublish {
		t.Fatalf("unexpected principal: %#v", result.Principal)
	}
	if result.Principal.PrincipalID != landlordID.Hex() || result.Principal.Phone != "13800000000" {
		t.Fatalf("unexpected principal identity: %#v", result.Principal)
	}
	if len(result.Principal.RoleCodes) != 0 || len(result.Principal.PermissionCodes) != 0 {
		t.Fatalf("expected no role or permission codes, got %#v", result.Principal)
	}
	if result.Subject.ID != landlordID.Hex() || result.Subject.Phone != "13800000000" {
		t.Fatalf("unexpected subject: %#v", result.Subject)
	}
	if landlordAuthRepo.touchedAuthID != authID || landlordAuthRepo.touchedLoginIP != "127.0.0.1" {
		t.Fatalf("expected login touch, got authID=%s ip=%s", landlordAuthRepo.touchedAuthID.Hex(), landlordAuthRepo.touchedLoginIP)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	landlordID := bson.NewObjectID()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	svc := newService(
		&fakeLandlordRepository{
			activeByPhone: &authmodel.Landlord{CommonFields: commonmodel.CommonFields{ID: landlordID}, Phone: "13800000000"},
			byID:          &authmodel.Landlord{CommonFields: commonmodel.CommonFields{ID: landlordID}, Phone: "13800000000"},
		},
		&fakeLandlordAuthRepository{activeByLandlordID: &authmodel.LandlordAuth{CommonFields: commonmodel.CommonFields{ID: bson.NewObjectID()}, LandlordID: landlordID, AuthType: authmodel.PasswordAuthTypePassword, PasswordHash: string(hash)}},
		session.NewStore(newFakeCache(), 0),
	)

	_, err = svc.Login(context.Background(), LoginInput{Phone: "13800000000", Password: "bad"})

	got := errcode.FromError(err)
	if got == nil || got.Code != errcode.Unauthorized.Code {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

type fakeLandlordRepository struct {
	activeByPhone *authmodel.Landlord
	byID          *authmodel.Landlord
}

func (f *fakeLandlordRepository) FindActiveByPhone(ctx context.Context, phone string) (*authmodel.Landlord, error) {
	return f.activeByPhone, nil
}

func (f *fakeLandlordRepository) FindActiveByID(ctx context.Context, id bson.ObjectID) (*authmodel.Landlord, error) {
	if f.byID != nil {
		return f.byID, nil
	}
	return f.activeByPhone, nil
}

type fakeLandlordAuthRepository struct {
	activeByLandlordID *authmodel.LandlordAuth
	touchedAuthID      bson.ObjectID
	touchedLoginIP     string
}

func (f *fakeLandlordAuthRepository) FindActivePasswordByLandlordID(ctx context.Context, landlordID bson.ObjectID) (*authmodel.LandlordAuth, error) {
	return f.activeByLandlordID, nil
}

func (f *fakeLandlordAuthRepository) TouchLastLogin(ctx context.Context, authID bson.ObjectID, loginIP string) error {
	f.touchedAuthID = authID
	f.touchedLoginIP = loginIP
	return nil
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
