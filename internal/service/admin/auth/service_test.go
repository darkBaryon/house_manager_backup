package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginRequiresPhoneAndPassword(t *testing.T) {
	svc := newService(nil, nil, nil, nil, nil, nil, nil, nil)

	_, err := svc.Login(context.Background(), LoginInput{})

	got := errcode.FromError(err)
	if got == nil || got.Code != errcode.InvalidParam.Code || got.PublicMessage() != "请输入手机号和密码" {
		t.Fatalf("expected invalid param with chinese message, got %v", err)
	}
}

func TestLoginBuildsStaffPrincipalAndPermissions(t *testing.T) {
	staffID := bson.NewObjectID()
	authID := bson.NewObjectID()
	roleID1 := bson.NewObjectID()
	roleID2 := bson.NewObjectID()
	permissionID1 := bson.NewObjectID()
	permissionID2 := bson.NewObjectID()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	loginLogRepo := &fakeLoginLogRepository{}
	staffAuthRepo := &fakeStaffAuthRepository{
		activeByStaffID: &authmodel.AdmStaffAuth{
			CommonFields: commonmodel.CommonFields{ID: authID, Status: commonmodel.StatusActive},
			StaffID:      staffID,
			AuthType:     authmodel.PasswordAuthTypePassword,
			PasswordHash: string(hash),
		},
	}
	svc := newService(
		&fakeStaffRepository{
			activeByPhone: &authmodel.AdmStaff{
				CommonFields:     commonmodel.CommonFields{ID: staffID, Status: commonmodel.StatusActive},
				Name:             "管理员",
				Phone:            "13800000000",
				Email:            "admin@example.com",
				Department:       "运营",
				JobTitle:         "主管",
				ContactQRCode:    "https://cdn.example.com/qr.png",
				CreatedByStaffID: bson.NewObjectID(),
			},
		},
		staffAuthRepo,
		&fakeStaffRoleRepository{
			items: []authmodel.AdmStaffRole{
				{RoleID: roleID2},
				{RoleID: roleID1},
			},
		},
		&fakeRoleRepository{
			items: []authmodel.AdmRole{
				{RoleCode: "viewer"},
				{RoleCode: "super_admin"},
			},
		},
		&fakeRolePermissionRepository{
			items: []authmodel.AdmRolePermission{
				{PermissionID: permissionID2},
				{PermissionID: permissionID1},
			},
		},
		&fakePermissionRepository{
			items: []authmodel.AdmPermission{
				{PermissionCode: "staff.edit"},
				{PermissionCode: "staff.view"},
			},
		},
		loginLogRepo,
		session.NewStore(newFakeCache(), 0),
	)

	result, err := svc.Login(context.Background(), LoginInput{
		Phone:     "13800000000",
		Password:  "secret123",
		LoginIP:   "127.0.0.1",
		UserAgent: "Mozilla/5.0",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected token")
	}
	if result.Principal.PrincipalType != session.PrincipalTypeStaff || result.Principal.Terminal != session.TerminalAdmin {
		t.Fatalf("unexpected principal: %#v", result.Principal)
	}
	if result.Principal.PrincipalID != staffID.Hex() || result.Principal.Phone != "13800000000" {
		t.Fatalf("unexpected principal identity: %#v", result.Principal)
	}
	if len(result.RoleCodes) != 2 || result.RoleCodes[0] != "super_admin" || result.RoleCodes[1] != "viewer" {
		t.Fatalf("unexpected role codes: %#v", result.RoleCodes)
	}
	if len(result.PermissionCodes) != 2 || result.PermissionCodes[0] != "staff.edit" || result.PermissionCodes[1] != "staff.view" {
		t.Fatalf("unexpected permission codes: %#v", result.PermissionCodes)
	}
	if result.StaffProfile.StaffID != staffID.Hex() || result.StaffProfile.Name != "管理员" || result.StaffProfile.ContactQRCode != "https://cdn.example.com/qr.png" {
		t.Fatalf("unexpected staff profile: %#v", result.StaffProfile)
	}
	if staffAuthRepo.touchedAuthID != authID || staffAuthRepo.touchedLoginIP != "127.0.0.1" {
		t.Fatalf("expected last login touch, got authID=%s ip=%s", staffAuthRepo.touchedAuthID.Hex(), staffAuthRepo.touchedLoginIP)
	}
	if loginLogRepo.entry == nil || loginLogRepo.entry.StaffID != staffID || loginLogRepo.entry.LoginResult != 1 || loginLogRepo.entry.UserAgent != "Mozilla/5.0" {
		t.Fatalf("unexpected login log entry: %#v", loginLogRepo.entry)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	staffID := bson.NewObjectID()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	svc := newService(
		&fakeStaffRepository{
			activeByPhone: &authmodel.AdmStaff{CommonFields: commonmodel.CommonFields{ID: staffID, Status: commonmodel.StatusActive}, Name: "管理员", Phone: "13800000000"},
		},
		&fakeStaffAuthRepository{
			activeByStaffID: &authmodel.AdmStaffAuth{CommonFields: commonmodel.CommonFields{ID: bson.NewObjectID(), Status: commonmodel.StatusActive}, StaffID: staffID, AuthType: authmodel.PasswordAuthTypePassword, PasswordHash: string(hash)},
		},
		&fakeStaffRoleRepository{},
		&fakeRoleRepository{},
		&fakeRolePermissionRepository{},
		&fakePermissionRepository{},
		nil,
		session.NewStore(newFakeCache(), 0),
	)

	_, err = svc.Login(context.Background(), LoginInput{Phone: "13800000000", Password: "bad"})

	got := errcode.FromError(err)
	if got == nil || got.Code != errcode.Unauthorized.Code || got.PublicMessage() != "手机号或密码错误，请重新输入" {
		t.Fatalf("expected unauthorized with chinese message, got %v", err)
	}
}

func TestLoginRejectsMissingStaffOrAuthRecord(t *testing.T) {
	t.Run("staff not found", func(t *testing.T) {
		svc := newService(
			&fakeStaffRepository{},
			&fakeStaffAuthRepository{},
			&fakeStaffRoleRepository{},
			&fakeRoleRepository{},
			&fakeRolePermissionRepository{},
			&fakePermissionRepository{},
			nil,
			session.NewStore(newFakeCache(), 0),
		)

		_, err := svc.Login(context.Background(), LoginInput{Phone: "13800000000", Password: "secret123"})

		assertErr(t, err, errcode.Unauthorized.Code, "手机号或密码错误，请重新输入")
	})

	t.Run("auth record not found", func(t *testing.T) {
		staffID := bson.NewObjectID()
		svc := newService(
			&fakeStaffRepository{
				activeByPhone: &authmodel.AdmStaff{CommonFields: commonmodel.CommonFields{ID: staffID, Status: commonmodel.StatusActive}, Name: "管理员", Phone: "13800000000"},
			},
			&fakeStaffAuthRepository{},
			&fakeStaffRoleRepository{},
			&fakeRoleRepository{},
			&fakeRolePermissionRepository{},
			&fakePermissionRepository{},
			nil,
			session.NewStore(newFakeCache(), 0),
		)

		_, err := svc.Login(context.Background(), LoginInput{Phone: "13800000000", Password: "secret123"})

		assertErr(t, err, errcode.Unauthorized.Code, "手机号或密码错误，请重新输入")
	})
}

func TestLoginDependencyFailuresReturnChineseMessages(t *testing.T) {
	staffID := bson.NewObjectID()
	authID := bson.NewObjectID()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	t.Run("staff repo failure", func(t *testing.T) {
		svc := newService(
			&fakeStaffRepository{activeByPhoneErr: fmt.Errorf("mongo down")},
			&fakeStaffAuthRepository{},
			&fakeStaffRoleRepository{},
			&fakeRoleRepository{},
			&fakeRolePermissionRepository{},
			&fakePermissionRepository{},
			nil,
			session.NewStore(newFakeCache(), 0),
		)

		_, err := svc.Login(context.Background(), LoginInput{Phone: "13800000000", Password: "secret123"})

		assertErr(t, err, errcode.DatabaseError.Code, "登录失败，请稍后重试")
	})

	t.Run("permission lookup failure", func(t *testing.T) {
		svc := newService(
			&fakeStaffRepository{
				activeByPhone: &authmodel.AdmStaff{CommonFields: commonmodel.CommonFields{ID: staffID, Status: commonmodel.StatusActive}, Name: "管理员", Phone: "13800000000"},
			},
			&fakeStaffAuthRepository{
				activeByStaffID: &authmodel.AdmStaffAuth{CommonFields: commonmodel.CommonFields{ID: authID}, StaffID: staffID, PasswordHash: string(hash)},
			},
			&fakeStaffRoleRepository{items: []authmodel.AdmStaffRole{{RoleID: bson.NewObjectID()}}},
			&fakeRoleRepository{items: []authmodel.AdmRole{{RoleCode: "super_admin"}}},
			&fakeRolePermissionRepository{items: []authmodel.AdmRolePermission{{PermissionID: bson.NewObjectID()}}},
			&fakePermissionRepository{err: fmt.Errorf("mongo down")},
			nil,
			session.NewStore(newFakeCache(), 0),
		)

		_, err := svc.Login(context.Background(), LoginInput{Phone: "13800000000", Password: "secret123"})

		assertErr(t, err, errcode.DatabaseError.Code, "登录失败，请稍后重试")
	})

	t.Run("session store failure", func(t *testing.T) {
		svc := newService(
			&fakeStaffRepository{
				activeByPhone: &authmodel.AdmStaff{CommonFields: commonmodel.CommonFields{ID: staffID, Status: commonmodel.StatusActive}, Name: "管理员", Phone: "13800000000"},
			},
			&fakeStaffAuthRepository{
				activeByStaffID: &authmodel.AdmStaffAuth{CommonFields: commonmodel.CommonFields{ID: authID}, StaffID: staffID, PasswordHash: string(hash)},
			},
			&fakeStaffRoleRepository{},
			&fakeRoleRepository{},
			&fakeRolePermissionRepository{},
			&fakePermissionRepository{},
			nil,
			session.NewStore(&fakeCache{values: map[string]string{}, setErr: fmt.Errorf("redis down")}, 0),
		)

		_, err := svc.Login(context.Background(), LoginInput{Phone: "13800000000", Password: "secret123"})

		assertErr(t, err, errcode.CacheError.Code, "登录失败，请稍后重试")
	})
}

func TestLoginNonBlockingAuditFailures(t *testing.T) {
	staffID := bson.NewObjectID()
	authID := bson.NewObjectID()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	svc := newService(
		&fakeStaffRepository{
			activeByPhone: &authmodel.AdmStaff{CommonFields: commonmodel.CommonFields{ID: staffID, Status: commonmodel.StatusActive}, Name: "管理员", Phone: "13800000000"},
		},
		&fakeStaffAuthRepository{
			activeByStaffID: &authmodel.AdmStaffAuth{CommonFields: commonmodel.CommonFields{ID: authID}, StaffID: staffID, PasswordHash: string(hash)},
			touchErr:        fmt.Errorf("touch failed"),
		},
		&fakeStaffRoleRepository{},
		&fakeRoleRepository{},
		&fakeRolePermissionRepository{},
		&fakePermissionRepository{},
		&fakeLoginLogRepository{err: fmt.Errorf("log failed")},
		session.NewStore(newFakeCache(), 0),
	)

	result, err := svc.Login(context.Background(), LoginInput{Phone: "13800000000", Password: "secret123"})

	if err != nil {
		t.Fatalf("expected login success despite audit failures, got %v", err)
	}
	if result == nil || result.Token == "" {
		t.Fatalf("expected token")
	}
}

func TestSessionReturnsCurrentStaffSession(t *testing.T) {
	staffID := bson.NewObjectID()
	svc := newService(
		&fakeStaffRepository{
			byID: &authmodel.AdmStaff{
				CommonFields:  commonmodel.CommonFields{ID: staffID, Status: commonmodel.StatusActive},
				Name:          "管理员",
				Phone:         "13800000000",
				Email:         "admin@example.com",
				Department:    "运营",
				JobTitle:      "主管",
				ContactQRCode: "https://cdn.example.com/qr.png",
			},
		},
		&fakeStaffAuthRepository{},
		&fakeStaffRoleRepository{items: []authmodel.AdmStaffRole{{RoleID: bson.NewObjectID()}}},
		&fakeRoleRepository{items: []authmodel.AdmRole{{RoleCode: "super_admin"}}},
		&fakeRolePermissionRepository{items: []authmodel.AdmRolePermission{{PermissionID: bson.NewObjectID()}}},
		&fakePermissionRepository{items: []authmodel.AdmPermission{{PermissionCode: "staff.view"}}},
		nil,
		session.NewStore(newFakeCache(), 0),
	)

	result, err := svc.Session(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeStaff,
		PrincipalID:   staffID.Hex(),
		Terminal:      session.TerminalAdmin,
		Phone:         "13800000000",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Principal.PrincipalType != session.PrincipalTypeStaff || result.Principal.Terminal != session.TerminalAdmin {
		t.Fatalf("unexpected principal: %#v", result.Principal)
	}
	if len(result.RoleCodes) != 1 || result.RoleCodes[0] != "super_admin" {
		t.Fatalf("unexpected role codes: %#v", result.RoleCodes)
	}
	if len(result.PermissionCodes) != 1 || result.PermissionCodes[0] != "staff.view" {
		t.Fatalf("unexpected permission codes: %#v", result.PermissionCodes)
	}
}

func TestSessionRejectsWrongPrincipal(t *testing.T) {
	svc := newService(nil, nil, nil, nil, nil, nil, nil, nil)

	_, err := svc.Session(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeLandlord,
		PrincipalID:   "staff-id",
		Terminal:      session.TerminalPublish,
	})

	got := errcode.FromError(err)
	if got == nil || got.Code != errcode.Unauthorized.Code || got.PublicMessage() != "当前登录状态无效，请重新登录" {
		t.Fatalf("expected unauthorized with chinese message, got %v", err)
	}
}

func TestSessionRejectsInvalidPrincipalIDAndMissingStaff(t *testing.T) {
	svc := newService(
		&fakeStaffRepository{},
		nil,
		&fakeStaffRoleRepository{},
		&fakeRoleRepository{},
		&fakeRolePermissionRepository{},
		&fakePermissionRepository{},
		nil,
		nil,
	)

	_, err := svc.Session(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeStaff,
		PrincipalID:   "bad-id",
		Terminal:      session.TerminalAdmin,
	})
	assertErr(t, err, errcode.Unauthorized.Code, "当前登录状态无效，请重新登录")

	_, err = svc.Session(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeStaff,
		PrincipalID:   bson.NewObjectID().Hex(),
		Terminal:      session.TerminalAdmin,
	})
	assertErr(t, err, errcode.Unauthorized.Code, "当前登录状态已失效，请重新登录")
}

func TestLogoutDeletesCurrentToken(t *testing.T) {
	cache := newFakeCache()
	store := session.NewStore(cache, 0)
	token, err := store.CreatePrincipal(context.Background(), session.Principal{
		PrincipalType: session.PrincipalTypeStaff,
		PrincipalID:   "staff-id",
		Terminal:      session.TerminalAdmin,
		Phone:         "13800000000",
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	svc := newService(nil, nil, nil, nil, nil, nil, nil, store)

	if err := svc.Logout(context.Background(), token); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cache.values["hs:sess:"+token]; ok {
		t.Fatalf("expected token deleted")
	}
}

func TestLogoutRejectsBlankTokenAndCacheFailure(t *testing.T) {
	svc := newService(nil, nil, nil, nil, nil, nil, nil, session.NewStore(newFakeCache(), 0))

	err := svc.Logout(context.Background(), "")

	assertErr(t, err, errcode.Unauthorized.Code, "未登录或登录已过期，请重新登录")

	svc = newService(nil, nil, nil, nil, nil, nil, nil, session.NewStore(&fakeCache{values: map[string]string{}, delErr: fmt.Errorf("redis down")}, 0))

	err = svc.Logout(context.Background(), "opaque-token")

	assertErr(t, err, errcode.CacheError.Code, "退出登录失败，请稍后重试")
}

func assertErr(t *testing.T, err error, code int, message string) {
	t.Helper()
	got := errcode.FromError(err)
	if got == nil || got.Code != code || got.PublicMessage() != message {
		t.Fatalf("expected code=%d message=%q, got %v", code, message, err)
	}
}

type fakeStaffRepository struct {
	activeByPhone    *authmodel.AdmStaff
	activeByPhoneErr error
	byID             *authmodel.AdmStaff
	findByIDErr      error
}

func (f *fakeStaffRepository) FindActiveByPhone(ctx context.Context, phone string) (*authmodel.AdmStaff, error) {
	if f.activeByPhoneErr != nil {
		return nil, f.activeByPhoneErr
	}
	return f.activeByPhone, nil
}

func (f *fakeStaffRepository) FindActiveByID(ctx context.Context, id bson.ObjectID) (*authmodel.AdmStaff, error) {
	if f.findByIDErr != nil {
		return nil, f.findByIDErr
	}
	if f.byID != nil {
		return f.byID, nil
	}
	return f.activeByPhone, nil
}

type fakeStaffAuthRepository struct {
	activeByStaffID *authmodel.AdmStaffAuth
	findErr         error
	touchedAuthID   bson.ObjectID
	touchedLoginIP  string
	touchErr        error
}

func (f *fakeStaffAuthRepository) FindActivePasswordByStaffID(ctx context.Context, staffID bson.ObjectID) (*authmodel.AdmStaffAuth, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.activeByStaffID, nil
}

func (f *fakeStaffAuthRepository) TouchLastLogin(ctx context.Context, authID bson.ObjectID, loginIP string) error {
	if f.touchErr != nil {
		return f.touchErr
	}
	f.touchedAuthID = authID
	f.touchedLoginIP = loginIP
	return nil
}

type fakeStaffRoleRepository struct {
	items []authmodel.AdmStaffRole
	err   error
}

func (f *fakeStaffRoleRepository) ListActiveByStaffID(ctx context.Context, staffID bson.ObjectID) ([]authmodel.AdmStaffRole, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

type fakeRoleRepository struct {
	items []authmodel.AdmRole
	err   error
}

func (f *fakeRoleRepository) FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]authmodel.AdmRole, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

type fakeRolePermissionRepository struct {
	items []authmodel.AdmRolePermission
	err   error
}

func (f *fakeRolePermissionRepository) ListActiveByRoleIDs(ctx context.Context, roleIDs []bson.ObjectID) ([]authmodel.AdmRolePermission, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

type fakePermissionRepository struct {
	items []authmodel.AdmPermission
	err   error
}

func (f *fakePermissionRepository) FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]authmodel.AdmPermission, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

type fakeLoginLogRepository struct {
	entry *authmodel.AdmLoginLog
	err   error
}

func (f *fakeLoginLogRepository) Create(ctx context.Context, logEntry *authmodel.AdmLoginLog) error {
	if f.err != nil {
		return f.err
	}
	f.entry = logEntry
	return nil
}

type fakeCache struct {
	values map[string]string
	setErr error
	delErr error
}

func newFakeCache() *fakeCache {
	return &fakeCache{values: map[string]string{}}
}

func (f *fakeCache) Get(ctx context.Context, key string) (string, error) {
	return f.values[key], nil
}

func (f *fakeCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if f.setErr != nil {
		return f.setErr
	}
	f.values[key] = value
	return nil
}

func (f *fakeCache) Del(ctx context.Context, keys ...string) error {
	if f.delErr != nil {
		return f.delErr
	}
	for _, key := range keys {
		delete(f.values, key)
	}
	return nil
}
