package role

import (
	"context"
	"errors"
	"testing"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	admrepo "house-manager/internal/repository/adm"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestCreateCreatesRoleAndPermissions(t *testing.T) {
	operatorID := bson.NewObjectID()
	permissionID := bson.NewObjectID()
	permission := authmodel.AdmPermission{PermissionCode: "staff.view", PermissionName: "查看员工", Module: "staff", Action: "view"}
	permission.ID = permissionID
	roleRepo := &fakeRoleRepository{}
	rolePermissionRepo := &fakeRolePermissionRepository{}
	svc := newService(roleRepo, &fakePermissionRepository{items: []authmodel.AdmPermission{permission}}, rolePermissionRepo, &fakeStaffRoleRepository{})

	result, err := svc.Create(context.Background(), CreateInput{
		OperatorStaffID: operatorID.Hex(),
		RoleName:        " 运营管理员 ",
		RoleCode:        " ops_admin ",
		Description:     " 运营 ",
		PermissionCodes: []string{"staff.view", "staff.view"},
	})
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	if result.Role.RoleName != "运营管理员" || result.Role.RoleCode != "ops_admin" {
		t.Fatalf("unexpected role: %#v", result.Role)
	}
	if roleRepo.created == nil || roleRepo.created.RoleCode != "ops_admin" {
		t.Fatalf("expected created role, got %#v", roleRepo.created)
	}
	if len(rolePermissionRepo.createdPermissionIDs) != 1 || rolePermissionRepo.createdPermissionIDs[0] != permissionID {
		t.Fatalf("unexpected permissions: %#v", rolePermissionRepo.createdPermissionIDs)
	}
	if rolePermissionRepo.assignedBy != operatorID {
		t.Fatalf("unexpected assigned by: %s", rolePermissionRepo.assignedBy.Hex())
	}
}

func TestCreateRejectsMissingPermission(t *testing.T) {
	svc := newService(&fakeRoleRepository{}, &fakePermissionRepository{}, &fakeRolePermissionRepository{}, &fakeStaffRoleRepository{})
	_, err := svc.Create(context.Background(), CreateInput{
		OperatorStaffID: bson.NewObjectID().Hex(),
		RoleName:        "运营管理员",
		RoleCode:        "ops_admin",
		PermissionCodes: []string{"staff.view"},
	})
	assertRoleErr(t, err, errcode.InvalidParam.Code, "所选权限不存在或已停用，请刷新后重试")
}

func TestCreateRollsBackRoleWhenPermissionRelationFails(t *testing.T) {
	permissionID := bson.NewObjectID()
	permission := authmodel.AdmPermission{PermissionCode: "staff.view", PermissionName: "查看员工"}
	permission.ID = permissionID
	roleRepo := &fakeRoleRepository{}
	rolePermissionRepo := &fakeRolePermissionRepository{err: errors.New("mongo down")}
	svc := newService(roleRepo, &fakePermissionRepository{items: []authmodel.AdmPermission{permission}}, rolePermissionRepo, &fakeStaffRoleRepository{})

	_, err := svc.Create(context.Background(), CreateInput{
		OperatorStaffID: bson.NewObjectID().Hex(),
		RoleName:        "运营管理员",
		RoleCode:        "ops_admin",
		PermissionCodes: []string{"staff.view"},
	})
	assertRoleErr(t, err, errcode.DatabaseError.Code, "创建角色失败，请稍后重试")
	if roleRepo.rollbackID != roleRepo.created.ID || rolePermissionRepo.rollbackRoleID != roleRepo.created.ID {
		t.Fatalf("expected rollback, role=%s permission=%s created=%s", roleRepo.rollbackID.Hex(), rolePermissionRepo.rollbackRoleID.Hex(), roleRepo.created.ID.Hex())
	}
}

func TestUpdateReplacesPermissionsAndInvalidatesAssignedStaff(t *testing.T) {
	operatorID := bson.NewObjectID()
	roleID := bson.NewObjectID()
	staffID := bson.NewObjectID()
	permissionID := bson.NewObjectID()
	role := &authmodel.AdmRole{RoleName: "运营管理员", RoleCode: "ops_admin"}
	role.ID = roleID
	role.Status = commonmodel.StatusActive
	permission := authmodel.AdmPermission{PermissionCode: "staff.edit", PermissionName: "编辑员工"}
	permission.ID = permissionID
	rolePermissionRepo := &fakeRolePermissionRepository{}
	sessionStore := &fakeSessionInvalidator{}
	svc := newService(
		&fakeRoleRepository{detail: role},
		&fakePermissionRepository{items: []authmodel.AdmPermission{permission}},
		rolePermissionRepo,
		&fakeStaffRoleRepository{items: []authmodel.AdmStaffRole{{RoleID: roleID, StaffID: staffID}}},
		sessionStore,
	)

	codes := []string{"staff.edit"}
	result, err := svc.Update(context.Background(), UpdateInput{
		OperatorStaffID: operatorID.Hex(),
		RoleID:          roleID.Hex(),
		PermissionCodes: &codes,
	})
	if err != nil {
		t.Fatalf("update role: %v", err)
	}
	if len(rolePermissionRepo.replacedPermissionIDs) != 1 || rolePermissionRepo.replacedPermissionIDs[0] != permissionID {
		t.Fatalf("unexpected replaced permissions: %#v", rolePermissionRepo.replacedPermissionIDs)
	}
	if rolePermissionRepo.replacedAssignedBy != operatorID {
		t.Fatalf("unexpected assigned by: %s", rolePermissionRepo.replacedAssignedBy.Hex())
	}
	if sessionStore.invalidatedPrincipalID != staffID.Hex() {
		t.Fatalf("expected assigned staff invalidated, got %#v", sessionStore)
	}
	if len(result.Role.PermissionCodes) != 1 || result.Role.PermissionCodes[0] != "staff.edit" {
		t.Fatalf("unexpected result: %#v", result.Role)
	}
}

func assertRoleErr(t *testing.T, err error, code int, message string) {
	t.Helper()
	got := errcode.FromError(err)
	if got == nil {
		t.Fatalf("expected errcode, got %v", err)
	}
	if got.Code != code || got.PublicMessage() != message {
		t.Fatalf("expected code=%d message=%q, got code=%d message=%q err=%v", code, message, got.Code, got.PublicMessage(), err)
	}
}

type fakeRoleRepository struct {
	created    *authmodel.AdmRole
	existing   *authmodel.AdmRole
	detail     *authmodel.AdmRole
	rollbackID bson.ObjectID
	listItems  []authmodel.AdmRole
	listTotal  int64
	err        error
}

func (f *fakeRoleRepository) Create(ctx context.Context, role *authmodel.AdmRole) error {
	if f.err != nil {
		return f.err
	}
	if role.ID.IsZero() {
		role.ID = bson.NewObjectID()
	}
	role.Status = commonmodel.StatusActive
	copied := *role
	f.created = &copied
	f.detail = &copied
	return nil
}

func (f *fakeRoleRepository) FindActiveByID(ctx context.Context, id bson.ObjectID) (*authmodel.AdmRole, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.detail, nil
}

func (f *fakeRoleRepository) FindByCode(ctx context.Context, roleCode string) (*authmodel.AdmRole, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.existing, nil
}

func (f *fakeRoleRepository) List(ctx context.Context, input admrepo.RoleListFilter) ([]authmodel.AdmRole, int64, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.listItems, f.listTotal, nil
}

func (f *fakeRoleRepository) Update(ctx context.Context, id bson.ObjectID, input admrepo.RoleUpdate) error {
	if f.err != nil {
		return f.err
	}
	if f.detail != nil {
		if input.RoleName != nil {
			f.detail.RoleName = *input.RoleName
		}
		if input.Description != nil {
			f.detail.Description = *input.Description
		}
	}
	return nil
}

func (f *fakeRoleRepository) RollbackCreate(ctx context.Context, id bson.ObjectID) error {
	f.rollbackID = id
	return nil
}

type fakePermissionRepository struct {
	items []authmodel.AdmPermission
	err   error
}

func (f *fakePermissionRepository) FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]authmodel.AdmPermission, error) {
	if f.err != nil {
		return nil, f.err
	}
	byID := make(map[bson.ObjectID]authmodel.AdmPermission, len(f.items))
	for _, item := range f.items {
		byID[item.ID] = item
	}
	result := make([]authmodel.AdmPermission, 0, len(ids))
	for _, id := range ids {
		if item, ok := byID[id]; ok {
			result = append(result, item)
		}
	}
	return result, nil
}

func (f *fakePermissionRepository) FindActiveByCodes(ctx context.Context, codes []string) ([]authmodel.AdmPermission, error) {
	if f.err != nil {
		return nil, f.err
	}
	byCode := make(map[string]authmodel.AdmPermission, len(f.items))
	for _, item := range f.items {
		byCode[item.PermissionCode] = item
	}
	result := make([]authmodel.AdmPermission, 0, len(codes))
	for _, code := range codes {
		if item, ok := byCode[code]; ok {
			result = append(result, item)
		}
	}
	return result, nil
}

type fakeRolePermissionRepository struct {
	createdPermissionIDs  []bson.ObjectID
	assignedBy            bson.ObjectID
	replacedPermissionIDs []bson.ObjectID
	replacedAssignedBy    bson.ObjectID
	rollbackRoleID        bson.ObjectID
	relations             []authmodel.AdmRolePermission
	err                   error
}

func (f *fakeRolePermissionRepository) CreateMany(ctx context.Context, roleID bson.ObjectID, permissionIDs []bson.ObjectID, assignedBy bson.ObjectID) error {
	if f.err != nil {
		return f.err
	}
	f.createdPermissionIDs = append([]bson.ObjectID(nil), permissionIDs...)
	f.assignedBy = assignedBy
	return nil
}

func (f *fakeRolePermissionRepository) ReplaceByRoleID(ctx context.Context, roleID bson.ObjectID, permissionIDs []bson.ObjectID, assignedBy bson.ObjectID) error {
	if f.err != nil {
		return f.err
	}
	f.replacedPermissionIDs = append([]bson.ObjectID(nil), permissionIDs...)
	f.replacedAssignedBy = assignedBy
	f.relations = f.relations[:0]
	for _, permissionID := range permissionIDs {
		f.relations = append(f.relations, authmodel.AdmRolePermission{RoleID: roleID, PermissionID: permissionID})
	}
	return nil
}

func (f *fakeRolePermissionRepository) ListActiveByRoleIDs(ctx context.Context, roleIDs []bson.ObjectID) ([]authmodel.AdmRolePermission, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.relations, nil
}

func (f *fakeRolePermissionRepository) RollbackCreateByRoleID(ctx context.Context, roleID bson.ObjectID) error {
	f.rollbackRoleID = roleID
	return nil
}

type fakeStaffRoleRepository struct {
	items []authmodel.AdmStaffRole
	err   error
}

func (f *fakeStaffRoleRepository) ListActiveByRoleID(ctx context.Context, roleID bson.ObjectID) ([]authmodel.AdmStaffRole, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

type fakeSessionInvalidator struct {
	invalidatedPrincipalID string
	err                    error
}

func (f *fakeSessionInvalidator) InvalidatePrincipal(ctx context.Context, principal session.Principal) error {
	if f.err != nil {
		return f.err
	}
	f.invalidatedPrincipalID = principal.PrincipalID
	return nil
}
