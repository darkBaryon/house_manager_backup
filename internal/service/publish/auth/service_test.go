package auth

import (
	"context"
	"testing"
	"time"

	"house-manager/internal/model"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestLoginRejectsPhoneOnlyOutsideLocal(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, nil, nil, nil, "prod")

	_, err := svc.Login(context.Background(), LoginInput{Phone: "13800000000"})

	got := errcode.FromError(err)
	if got == nil || got.Code != errcode.Forbidden.Code {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestLoginRequiresPhone(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, nil, nil, nil, "local")

	_, err := svc.Login(context.Background(), LoginInput{})

	got := errcode.FromError(err)
	if got == nil || got.Code != errcode.InvalidParam.Code {
		t.Fatalf("expected invalid param, got %v", err)
	}
}

func TestSessionRejectsNonPublishPrincipal(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, nil, nil, nil, "local")

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

func TestLoginBuildsStaffPrincipalWithRolesAndPermissions(t *testing.T) {
	staffID := bson.NewObjectID()
	roleID := bson.NewObjectID()
	permissionID := bson.NewObjectID()
	store := session.NewStore(newFakeCache(), 0)
	staffRepo := &fakeStaffRepository{
		activeByPhone: &model.AdmStaff{
			CommonFields: model.CommonFields{ID: staffID, Status: model.StatusActive},
			Name:         "管家",
			Phone:        "13800000000",
		},
	}
	svc := newService(
		staffRepo,
		&fakeRoleRepository{roles: []model.AdmRole{{
			CommonFields: model.CommonFields{ID: roleID, Status: model.StatusActive},
			RoleCode:     "super_admin",
		}}},
		&fakePermissionRepository{permissions: []model.AdmPermission{{
			CommonFields:   model.CommonFields{ID: permissionID, Status: model.StatusActive},
			PermissionCode: "house.manage",
		}}},
		&fakeStaffRoleRepository{staffRoles: []model.AdmStaffRole{{RoleID: roleID}}},
		&fakeRolePermissionRepository{rolePermissions: []model.AdmRolePermission{{PermissionID: permissionID}}},
		&fakeLoginLogRepository{},
		store,
		"local",
	)

	result, err := svc.Login(context.Background(), LoginInput{Phone: "13800000000", LoginIP: "127.0.0.1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected token")
	}
	if result.Principal.PrincipalType != session.PrincipalTypeStaff || result.Principal.Terminal != session.TerminalPublish {
		t.Fatalf("unexpected principal: %#v", result.Principal)
	}
	if result.Principal.PrincipalID != staffID.Hex() || result.Principal.Phone != "13800000000" {
		t.Fatalf("unexpected principal identity: %#v", result.Principal)
	}
	if len(result.Principal.RoleCodes) != 1 || result.Principal.RoleCodes[0] != "super_admin" {
		t.Fatalf("unexpected role codes: %#v", result.Principal.RoleCodes)
	}
	if len(result.Principal.PermissionCodes) != 1 || result.Principal.PermissionCodes[0] != "house.manage" {
		t.Fatalf("unexpected permission codes: %#v", result.Principal.PermissionCodes)
	}
	if !staffRepo.touched {
		t.Fatal("expected staff last login touched")
	}
}

type fakeStaffRepository struct {
	activeByPhone *model.AdmStaff
	byID          *model.AdmStaff
	touched       bool
}

func (f *fakeStaffRepository) FindActiveByPhone(ctx context.Context, phone string) (*model.AdmStaff, error) {
	return f.activeByPhone, nil
}

func (f *fakeStaffRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.AdmStaff, error) {
	if f.byID != nil {
		return f.byID, nil
	}
	return f.activeByPhone, nil
}

func (f *fakeStaffRepository) TouchLastLogin(ctx context.Context, staffID bson.ObjectID, lastLoginAt int64, lastLoginIP string) error {
	f.touched = true
	return nil
}

type fakeRoleRepository struct {
	roles []model.AdmRole
}

func (f *fakeRoleRepository) FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]model.AdmRole, error) {
	return f.roles, nil
}

type fakePermissionRepository struct {
	permissions []model.AdmPermission
}

func (f *fakePermissionRepository) FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]model.AdmPermission, error) {
	return f.permissions, nil
}

type fakeStaffRoleRepository struct {
	staffRoles []model.AdmStaffRole
}

func (f *fakeStaffRoleRepository) ListActiveByStaffID(ctx context.Context, staffID bson.ObjectID) ([]model.AdmStaffRole, error) {
	return f.staffRoles, nil
}

type fakeRolePermissionRepository struct {
	rolePermissions []model.AdmRolePermission
}

func (f *fakeRolePermissionRepository) ListActiveByRoleIDs(ctx context.Context, roleIDs []bson.ObjectID) ([]model.AdmRolePermission, error) {
	return f.rolePermissions, nil
}

type fakeLoginLogRepository struct{}

func (f *fakeLoginLogRepository) Create(ctx context.Context, log *model.AdmLoginLog) error {
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
