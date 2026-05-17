package staff

import (
	"context"
	"errors"
	"fmt"
	"testing"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	admrepo "house-manager/internal/repository/adm"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

func TestCreateCreatesStaffAuthAndRoles(t *testing.T) {
	operatorID := bson.NewObjectID()
	roleID := bson.NewObjectID()
	role := authmodel.AdmRole{
		RoleName: "运营管理员",
		RoleCode: "ops_admin",
	}
	role.ID = roleID
	role.Status = commonmodel.StatusActive

	staffRepo := &fakeStaffCreateRepository{}
	authRepo := &fakeStaffAuthCreateRepository{}
	staffRoleRepo := &fakeStaffRoleCreateRepository{}
	roleRepo := &fakeRoleCreateRepository{roles: []authmodel.AdmRole{role}}
	svc := newService(staffRepo, authRepo, staffRoleRepo, roleRepo)

	result, err := svc.Create(context.Background(), CreateInput{
		OperatorStaffID: operatorID.Hex(),
		Name:            "  运营一号 ",
		Phone:           "13800000000",
		Password:        " secret123 ",
		Email:           " ops@example.com ",
		Department:      " 运营部 ",
		JobTitle:        " 运营 ",
		ContactQRCode:   " https://example.com/qr.png ",
		RoleIDs:         []string{roleID.Hex(), roleID.Hex()},
	})
	if err != nil {
		t.Fatalf("create staff: %v", err)
	}
	if result.Staff.Name != "运营一号" || result.Staff.Phone != "13800000000" {
		t.Fatalf("unexpected staff summary: %#v", result.Staff)
	}
	if len(result.Staff.Roles) != 1 || result.Staff.Roles[0].RoleCode != "ops_admin" {
		t.Fatalf("unexpected roles: %#v", result.Staff.Roles)
	}
	if staffRepo.created == nil || staffRepo.created.CreatedByStaffID != operatorID {
		t.Fatalf("expected created staff with operator, got %#v", staffRepo.created)
	}
	if authRepo.created == nil || authRepo.created.StaffID != staffRepo.created.ID {
		t.Fatalf("expected auth created for staff, got %#v", authRepo.created)
	}
	if bcrypt.CompareHashAndPassword([]byte(authRepo.created.PasswordHash), []byte("secret123")) != nil {
		t.Fatalf("expected bcrypt hash for password")
	}
	if len(staffRoleRepo.createdRoleIDs) != 1 || staffRoleRepo.createdRoleIDs[0] != roleID {
		t.Fatalf("unexpected staff roles: %#v", staffRoleRepo.createdRoleIDs)
	}
	if staffRoleRepo.assignedBy != operatorID {
		t.Fatalf("expected assigned by operator, got %s", staffRoleRepo.assignedBy.Hex())
	}
}

func TestCreateRejectsInvalidInput(t *testing.T) {
	svc := newService(&fakeStaffCreateRepository{}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})
	_, err := svc.Create(context.Background(), CreateInput{
		OperatorStaffID: bson.NewObjectID().Hex(),
		Name:            "运营一号",
		Phone:           "10086",
		Password:        "secret123",
	})
	assertStaffCreateErr(t, err, errcode.InvalidParam.Code, "请输入正确的员工手机号")
}

func TestCreateRejectsDuplicatePhone(t *testing.T) {
	existing := &authmodel.AdmStaff{Phone: "13800000000"}
	existing.ID = bson.NewObjectID()
	svc := newService(&fakeStaffCreateRepository{existing: existing}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})

	_, err := svc.Create(context.Background(), CreateInput{
		OperatorStaffID: bson.NewObjectID().Hex(),
		Name:            "运营一号",
		Phone:           "13800000000",
		Password:        "secret123",
	})
	assertStaffCreateErr(t, err, errcode.AlreadyExists.Code, "员工手机号已存在，请更换后重试")
}

func TestCreateRejectsMissingRole(t *testing.T) {
	svc := newService(&fakeStaffCreateRepository{}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})

	_, err := svc.Create(context.Background(), CreateInput{
		OperatorStaffID: bson.NewObjectID().Hex(),
		Name:            "运营一号",
		Phone:           "13800000000",
		Password:        "secret123",
		RoleIDs:         []string{bson.NewObjectID().Hex()},
	})
	assertStaffCreateErr(t, err, errcode.InvalidParam.Code, "所选角色不存在或已停用，请刷新后重试")
}

func TestCreateMapsRepositoryFailure(t *testing.T) {
	svc := newService(
		&fakeStaffCreateRepository{createErr: errors.New("mongo down")},
		&fakeStaffAuthCreateRepository{},
		&fakeStaffRoleCreateRepository{},
		&fakeRoleCreateRepository{},
	)

	_, err := svc.Create(context.Background(), CreateInput{
		OperatorStaffID: bson.NewObjectID().Hex(),
		Name:            "运营一号",
		Phone:           "13800000000",
		Password:        "secret123",
	})
	assertStaffCreateErr(t, err, errcode.DatabaseError.Code, "创建员工失败，请稍后重试")
}

func TestCreateRollsBackStaffWhenAuthCreateFails(t *testing.T) {
	staffRepo := &fakeStaffCreateRepository{}
	authRepo := &fakeStaffAuthCreateRepository{err: errors.New("auth down")}
	staffRoleRepo := &fakeStaffRoleCreateRepository{}
	svc := newService(staffRepo, authRepo, staffRoleRepo, &fakeRoleCreateRepository{})

	_, err := svc.Create(context.Background(), CreateInput{
		OperatorStaffID: bson.NewObjectID().Hex(),
		Name:            "运营一号",
		Phone:           "13800000000",
		Password:        "secret123",
	})
	assertStaffCreateErr(t, err, errcode.DatabaseError.Code, "创建员工失败，请稍后重试")
	if staffRepo.rollbackID.IsZero() || staffRepo.rollbackID != staffRepo.created.ID {
		t.Fatalf("expected created staff rollback, created=%#v rollback=%s", staffRepo.created, staffRepo.rollbackID.Hex())
	}
	if authRepo.rollbackStaffID != staffRepo.created.ID || staffRoleRepo.rollbackStaffID != staffRepo.created.ID {
		t.Fatalf("expected auth/role rollback, auth=%s role=%s staff=%s", authRepo.rollbackStaffID.Hex(), staffRoleRepo.rollbackStaffID.Hex(), staffRepo.created.ID.Hex())
	}
}

func TestCreateRollsBackStaffAuthAndRolesWhenRoleCreateFails(t *testing.T) {
	roleID := bson.NewObjectID()
	role := authmodel.AdmRole{RoleName: "运营管理员", RoleCode: "ops_admin"}
	role.ID = roleID
	staffRepo := &fakeStaffCreateRepository{}
	authRepo := &fakeStaffAuthCreateRepository{}
	staffRoleRepo := &fakeStaffRoleCreateRepository{err: errors.New("role down")}
	svc := newService(staffRepo, authRepo, staffRoleRepo, &fakeRoleCreateRepository{roles: []authmodel.AdmRole{role}})

	_, err := svc.Create(context.Background(), CreateInput{
		OperatorStaffID: bson.NewObjectID().Hex(),
		Name:            "运营一号",
		Phone:           "13800000000",
		Password:        "secret123",
		RoleIDs:         []string{roleID.Hex()},
	})
	assertStaffCreateErr(t, err, errcode.DatabaseError.Code, "创建员工失败，请稍后重试")
	if staffRepo.rollbackID != staffRepo.created.ID || authRepo.rollbackStaffID != staffRepo.created.ID || staffRoleRepo.rollbackStaffID != staffRepo.created.ID {
		t.Fatalf("expected rollback for staff/auth/roles, staff=%s auth=%s roles=%s created=%s", staffRepo.rollbackID.Hex(), authRepo.rollbackStaffID.Hex(), staffRoleRepo.rollbackStaffID.Hex(), staffRepo.created.ID.Hex())
	}
}

func TestListReturnsStaffWithRolesAndLoginSummary(t *testing.T) {
	staffID := bson.NewObjectID()
	roleID := bson.NewObjectID()
	role := authmodel.AdmRole{RoleName: "运营管理员", RoleCode: "ops_admin"}
	role.ID = roleID
	staffRepo := &fakeStaffCreateRepository{
		listItems: []authmodel.AdmStaff{{
			Name:       "运营一号",
			Phone:      "13800000000",
			Department: "运营部",
			JobTitle:   "运营",
		}},
		listTotal: 1,
	}
	staffRepo.listItems[0].ID = staffID
	staffRepo.listItems[0].Status = commonmodel.StatusActive
	staffRepo.listItems[0].CreatedAt = 111
	staffRepo.listItems[0].UpdatedAt = 222
	authRepo := &fakeStaffAuthCreateRepository{
		auths: []authmodel.AdmStaffAuth{{
			StaffID:     staffID,
			LastLoginAt: 333,
			LastLoginIP: "127.0.0.1",
		}},
	}
	staffRoleRepo := &fakeStaffRoleCreateRepository{
		rolesByStaffIDs: []authmodel.AdmStaffRole{{StaffID: staffID, RoleID: roleID}},
	}
	roleRepo := &fakeRoleCreateRepository{roles: []authmodel.AdmRole{role}}
	svc := newService(staffRepo, authRepo, staffRoleRepo, roleRepo)

	result, err := svc.List(context.Background(), ListInput{Keyword: " 运营 ", Page: -1, PageSize: 999})
	if err != nil {
		t.Fatalf("list staff: %v", err)
	}
	if result.Page != 1 || result.PageSize != 100 || result.Total != 1 {
		t.Fatalf("unexpected page result: %#v", result)
	}
	if staffRepo.listInput.Keyword != "运营" || staffRepo.listInput.Status != commonmodel.StatusActive {
		t.Fatalf("unexpected list input: %#v", staffRepo.listInput)
	}
	if len(result.List) != 1 {
		t.Fatalf("expected one staff, got %#v", result.List)
	}
	item := result.List[0]
	if item.StaffID != staffID.Hex() || item.LastLoginAt != 333 || item.LastLoginIP != "127.0.0.1" {
		t.Fatalf("unexpected list item: %#v", item)
	}
	if len(item.Roles) != 1 || item.Roles[0].RoleCode != "ops_admin" {
		t.Fatalf("unexpected roles: %#v", item.Roles)
	}
}

func TestListFiltersByRoleID(t *testing.T) {
	staffID := bson.NewObjectID()
	roleID := bson.NewObjectID()
	staffRepo := &fakeStaffCreateRepository{listItems: []authmodel.AdmStaff{}, listTotal: 0}
	staffRoleRepo := &fakeStaffRoleCreateRepository{
		rolesByRoleID: []authmodel.AdmStaffRole{{StaffID: staffID, RoleID: roleID}},
	}
	svc := newService(staffRepo, &fakeStaffAuthCreateRepository{}, staffRoleRepo, &fakeRoleCreateRepository{})

	_, err := svc.List(context.Background(), ListInput{RoleID: roleID.Hex()})
	if err != nil {
		t.Fatalf("list staff: %v", err)
	}
	if len(staffRepo.listInput.StaffIDs) != 1 || staffRepo.listInput.StaffIDs[0] != staffID {
		t.Fatalf("expected role staff ids passed to list, got %#v", staffRepo.listInput.StaffIDs)
	}
}

func TestListRejectsInvalidStatus(t *testing.T) {
	status := 2
	svc := newService(&fakeStaffCreateRepository{}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})
	_, err := svc.List(context.Background(), ListInput{Status: &status})
	assertStaffCreateErr(t, err, errcode.InvalidParam.Code, "员工状态参数不正确")
}

func TestListRejectsInvalidRoleID(t *testing.T) {
	svc := newService(&fakeStaffCreateRepository{}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})
	_, err := svc.List(context.Background(), ListInput{RoleID: "bad-role"})
	assertStaffCreateErr(t, err, errcode.InvalidParam.Code, "角色参数不正确")
}

func TestListMapsRepositoryFailure(t *testing.T) {
	svc := newService(&fakeStaffCreateRepository{listErr: errors.New("mongo down")}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})
	_, err := svc.List(context.Background(), ListInput{})
	assertStaffCreateErr(t, err, errcode.DatabaseError.Code, "获取员工列表失败，请稍后重试")
}

func TestDetailReturnsStaffDetail(t *testing.T) {
	staffID := bson.NewObjectID()
	creatorID := bson.NewObjectID()
	roleID := bson.NewObjectID()
	role := authmodel.AdmRole{RoleName: "运营管理员", RoleCode: "ops_admin"}
	role.ID = roleID
	staff := &authmodel.AdmStaff{
		Name:             "运营一号",
		Phone:            "13800000000",
		Email:            "ops@example.com",
		Department:       "运营部",
		JobTitle:         "运营",
		CreatedByStaffID: creatorID,
	}
	staff.ID = staffID
	staff.Status = commonmodel.StatusDeleted
	staff.CreatedAt = 111
	staff.UpdatedAt = 222
	svc := newService(
		&fakeStaffCreateRepository{detail: staff},
		&fakeStaffAuthCreateRepository{auths: []authmodel.AdmStaffAuth{{
			StaffID:           staffID,
			PasswordUpdatedAt: 333,
			LastLoginAt:       444,
			LastLoginIP:       "127.0.0.1",
		}}},
		&fakeStaffRoleCreateRepository{rolesByStaffIDs: []authmodel.AdmStaffRole{{StaffID: staffID, RoleID: roleID}}},
		&fakeRoleCreateRepository{roles: []authmodel.AdmRole{role}},
	)

	result, err := svc.Detail(context.Background(), DetailInput{StaffID: staffID.Hex()})
	if err != nil {
		t.Fatalf("detail staff: %v", err)
	}
	if result.Staff.StaffID != staffID.Hex() || result.Staff.CreatedByStaffID != creatorID.Hex() {
		t.Fatalf("unexpected staff detail: %#v", result.Staff)
	}
	if result.Staff.Status != commonmodel.StatusDeleted || result.Staff.PasswordUpdatedAt != 333 || result.Staff.LastLoginAt != 444 {
		t.Fatalf("unexpected status/auth detail: %#v", result.Staff)
	}
	if len(result.Staff.Roles) != 1 || result.Staff.Roles[0].RoleCode != "ops_admin" {
		t.Fatalf("unexpected roles: %#v", result.Staff.Roles)
	}
}

func TestDetailRejectsInvalidStaffID(t *testing.T) {
	svc := newService(&fakeStaffCreateRepository{}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})
	_, err := svc.Detail(context.Background(), DetailInput{StaffID: "bad-staff"})
	assertStaffCreateErr(t, err, errcode.InvalidParam.Code, "员工参数不正确")
}

func TestDetailReturnsNotFound(t *testing.T) {
	svc := newService(&fakeStaffCreateRepository{}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})
	_, err := svc.Detail(context.Background(), DetailInput{StaffID: bson.NewObjectID().Hex()})
	assertStaffCreateErr(t, err, errcode.NotFound.Code, "员工不存在或已删除")
}

func TestDetailMapsRepositoryFailure(t *testing.T) {
	svc := newService(&fakeStaffCreateRepository{detailErr: errors.New("mongo down")}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})
	_, err := svc.Detail(context.Background(), DetailInput{StaffID: bson.NewObjectID().Hex()})
	assertStaffCreateErr(t, err, errcode.DatabaseError.Code, "获取员工详情失败，请稍后重试")
}

func TestUpdateChangesFieldsAndRoles(t *testing.T) {
	operatorID := bson.NewObjectID()
	staffID := bson.NewObjectID()
	roleID := bson.NewObjectID()
	role := authmodel.AdmRole{RoleName: "运营管理员", RoleCode: "ops_admin"}
	role.ID = roleID
	staff := &authmodel.AdmStaff{
		Name:       "运营一号",
		Phone:      "13800000000",
		Department: "运营部",
	}
	staff.ID = staffID
	staff.Status = commonmodel.StatusActive
	staff.CreatedAt = 111
	staff.UpdatedAt = 222
	staffRepo := &fakeStaffCreateRepository{detail: staff}
	staffRoleRepo := &fakeStaffRoleCreateRepository{
		rolesByStaffIDs: []authmodel.AdmStaffRole{{StaffID: staffID, RoleID: roleID}},
	}
	sessionStore := &fakeSessionInvalidator{}
	svc := newService(staffRepo, &fakeStaffAuthCreateRepository{}, staffRoleRepo, &fakeRoleCreateRepository{roles: []authmodel.AdmRole{role}}, sessionStore)

	name := " 运营二号 "
	email := " ops2@example.com "
	status := commonmodel.StatusDeleted
	roleIDs := []string{roleID.Hex(), roleID.Hex()}
	result, err := svc.Update(context.Background(), UpdateInput{
		OperatorStaffID: operatorID.Hex(),
		StaffID:         staffID.Hex(),
		Name:            &name,
		Email:           &email,
		Status:          &status,
		RoleIDs:         &roleIDs,
	})
	if err != nil {
		t.Fatalf("update staff: %v", err)
	}
	if staffRepo.updatedID != staffID || staffRepo.updatedFields["name"] != "运营二号" || staffRepo.updatedFields["email"] != "ops2@example.com" {
		t.Fatalf("unexpected update fields: id=%s fields=%#v", staffRepo.updatedID.Hex(), staffRepo.updatedFields)
	}
	if staffRepo.updatedFields["status"] != commonmodel.StatusDeleted || result.Staff.Status != commonmodel.StatusDeleted {
		t.Fatalf("expected deleted status, fields=%#v result=%#v", staffRepo.updatedFields, result.Staff)
	}
	if len(staffRoleRepo.replacedRoleIDs) != 1 || staffRoleRepo.replacedRoleIDs[0] != roleID || staffRoleRepo.replacedAssignedBy != operatorID {
		t.Fatalf("unexpected replaced roles: %s", staffRoleRepo)
	}
	if sessionStore.invalidatedPrincipalID != staffID.Hex() {
		t.Fatalf("expected staff session invalidated, got %#v", sessionStore)
	}
	if result.Staff.Name != "运营二号" || len(result.Staff.Roles) != 1 || result.Staff.Roles[0].RoleCode != "ops_admin" {
		t.Fatalf("unexpected update result: %#v", result.Staff)
	}
}

func TestUpdateRejectsNoFields(t *testing.T) {
	svc := newService(&fakeStaffCreateRepository{}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})
	_, err := svc.Update(context.Background(), UpdateInput{StaffID: bson.NewObjectID().Hex()})
	assertStaffCreateErr(t, err, errcode.InvalidParam.Code, "请至少提交一个需要修改的字段")
}

func TestUpdateRejectsBlankName(t *testing.T) {
	name := "   "
	svc := newService(&fakeStaffCreateRepository{}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})
	_, err := svc.Update(context.Background(), UpdateInput{StaffID: bson.NewObjectID().Hex(), Name: &name})
	assertStaffCreateErr(t, err, errcode.InvalidParam.Code, "请输入员工姓名")
}

func TestUpdateRejectsMissingRole(t *testing.T) {
	staffID := bson.NewObjectID()
	staff := &authmodel.AdmStaff{Name: "运营一号", Phone: "13800000000"}
	staff.ID = staffID
	roleIDs := []string{bson.NewObjectID().Hex()}
	svc := newService(&fakeStaffCreateRepository{detail: staff}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})

	_, err := svc.Update(context.Background(), UpdateInput{
		OperatorStaffID: bson.NewObjectID().Hex(),
		StaffID:         staffID.Hex(),
		RoleIDs:         &roleIDs,
	})
	assertStaffCreateErr(t, err, errcode.InvalidParam.Code, "所选角色不存在或已停用，请刷新后重试")
}

func TestUpdateReturnsNotFound(t *testing.T) {
	name := "运营二号"
	svc := newService(&fakeStaffCreateRepository{}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})
	_, err := svc.Update(context.Background(), UpdateInput{StaffID: bson.NewObjectID().Hex(), Name: &name})
	assertStaffCreateErr(t, err, errcode.NotFound.Code, "员工不存在或已删除")
}

func TestDisableSetsStaffDeleted(t *testing.T) {
	operatorID := bson.NewObjectID()
	staffID := bson.NewObjectID()
	staff := &authmodel.AdmStaff{Name: "运营一号", Phone: "13800000000"}
	staff.ID = staffID
	staff.Status = commonmodel.StatusActive
	staffRepo := &fakeStaffCreateRepository{detail: staff}
	sessionStore := &fakeSessionInvalidator{}
	svc := newService(staffRepo, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{}, sessionStore)

	result, err := svc.Disable(context.Background(), DisableInput{
		OperatorStaffID: operatorID.Hex(),
		StaffID:         staffID.Hex(),
	})
	if err != nil {
		t.Fatalf("disable staff: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success")
	}
	if staffRepo.updatedFields["status"] != commonmodel.StatusDeleted || staff.Status != commonmodel.StatusDeleted {
		t.Fatalf("expected deleted status, fields=%#v staff=%#v", staffRepo.updatedFields, staff)
	}
	if sessionStore.invalidatedPrincipalID != staffID.Hex() {
		t.Fatalf("expected staff session invalidated, got %#v", sessionStore)
	}
}

func TestDisableReturnsNotFound(t *testing.T) {
	svc := newService(&fakeStaffCreateRepository{}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})
	_, err := svc.Disable(context.Background(), DisableInput{
		OperatorStaffID: bson.NewObjectID().Hex(),
		StaffID:         bson.NewObjectID().Hex(),
	})
	assertStaffCreateErr(t, err, errcode.NotFound.Code, "员工不存在或已删除")
}

func TestDisableRejectsSelf(t *testing.T) {
	staffID := bson.NewObjectID()
	svc := newService(&fakeStaffCreateRepository{}, &fakeStaffAuthCreateRepository{}, &fakeStaffRoleCreateRepository{}, &fakeRoleCreateRepository{})
	_, err := svc.Disable(context.Background(), DisableInput{
		OperatorStaffID: staffID.Hex(),
		StaffID:         staffID.Hex(),
	})
	assertStaffCreateErr(t, err, errcode.InvalidParam.Code, "不能禁用当前登录账号")
}

func assertStaffCreateErr(t *testing.T, err error, code int, message string) {
	t.Helper()
	got := errcode.FromError(err)
	if got == nil {
		t.Fatalf("expected errcode, got %v", err)
	}
	if got.Code != code || got.PublicMessage() != message {
		t.Fatalf("expected code=%d message=%q, got code=%d message=%q err=%v", code, message, got.Code, got.PublicMessage(), err)
	}
}

type fakeStaffCreateRepository struct {
	existing      *authmodel.AdmStaff
	created       *authmodel.AdmStaff
	createErr     error
	detail        *authmodel.AdmStaff
	detailErr     error
	updatedID     bson.ObjectID
	updatedFields bson.M
	updateErr     error
	rollbackID    bson.ObjectID
	rollbackErr   error
	listInput     admrepo.StaffListFilter
	listItems     []authmodel.AdmStaff
	listTotal     int64
	listErr       error
}

func (f *fakeStaffCreateRepository) Create(ctx context.Context, staff *authmodel.AdmStaff) error {
	if f.createErr != nil {
		return f.createErr
	}
	if staff.ID.IsZero() {
		staff.ID = bson.NewObjectID()
	}
	if staff.Status == 0 {
		staff.Status = commonmodel.StatusActive
	}
	if staff.CreatedAt == 0 {
		staff.CreatedAt = 111
	}
	if staff.UpdatedAt == 0 {
		staff.UpdatedAt = 222
	}
	copied := *staff
	f.created = &copied
	return nil
}

func (f *fakeStaffCreateRepository) FindByPhone(ctx context.Context, phone string) (*authmodel.AdmStaff, error) {
	return f.existing, nil
}

func (f *fakeStaffCreateRepository) FindByID(ctx context.Context, id bson.ObjectID) (*authmodel.AdmStaff, error) {
	if f.detailErr != nil {
		return nil, f.detailErr
	}
	return f.detail, nil
}

func (f *fakeStaffCreateRepository) List(ctx context.Context, input admrepo.StaffListFilter) ([]authmodel.AdmStaff, int64, error) {
	f.listInput = input
	if f.listErr != nil {
		return nil, 0, f.listErr
	}
	return f.listItems, f.listTotal, nil
}

func (f *fakeStaffCreateRepository) UpdateFields(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	f.updatedID = id
	f.updatedFields = bson.M{}
	for key, value := range fields {
		f.updatedFields[key] = value
	}
	if f.updateErr != nil {
		return f.updateErr
	}
	if f.detail != nil && f.detail.ID == id {
		if value, ok := fields["name"].(string); ok {
			f.detail.Name = value
		}
		if value, ok := fields["email"].(string); ok {
			f.detail.Email = value
		}
		if value, ok := fields["department"].(string); ok {
			f.detail.Department = value
		}
		if value, ok := fields["job_title"].(string); ok {
			f.detail.JobTitle = value
		}
		if value, ok := fields["contact_qr_code"].(string); ok {
			f.detail.ContactQRCode = value
		}
		if value, ok := fields["status"].(int); ok {
			f.detail.Status = value
		}
	}
	return nil
}

func (f *fakeStaffCreateRepository) RollbackCreate(ctx context.Context, id bson.ObjectID) error {
	if f.rollbackErr != nil {
		return f.rollbackErr
	}
	f.rollbackID = id
	return nil
}

type fakeStaffAuthCreateRepository struct {
	created         *authmodel.AdmStaffAuth
	rollbackStaffID bson.ObjectID
	auths           []authmodel.AdmStaffAuth
	err             error
}

func (f *fakeStaffAuthCreateRepository) CreatePasswordAuth(ctx context.Context, authRecord *authmodel.AdmStaffAuth) error {
	if f.err != nil {
		return f.err
	}
	copied := *authRecord
	f.created = &copied
	return nil
}

func (f *fakeStaffAuthCreateRepository) FindActivePasswordByStaffIDs(ctx context.Context, staffIDs []bson.ObjectID) ([]authmodel.AdmStaffAuth, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.auths, nil
}

func (f *fakeStaffAuthCreateRepository) RollbackCreateByStaffID(ctx context.Context, staffID bson.ObjectID) error {
	f.rollbackStaffID = staffID
	return nil
}

type fakeStaffRoleCreateRepository struct {
	createdStaffID     bson.ObjectID
	createdRoleIDs     []bson.ObjectID
	assignedBy         bson.ObjectID
	rolesByRoleID      []authmodel.AdmStaffRole
	rolesByStaffIDs    []authmodel.AdmStaffRole
	replacedStaffID    bson.ObjectID
	replacedRoleIDs    []bson.ObjectID
	replacedAssignedBy bson.ObjectID
	rollbackStaffID    bson.ObjectID
	err                error
}

func (f *fakeStaffRoleCreateRepository) CreateMany(ctx context.Context, staffID bson.ObjectID, roleIDs []bson.ObjectID, assignedBy bson.ObjectID) error {
	if f.err != nil {
		return f.err
	}
	f.createdStaffID = staffID
	f.createdRoleIDs = append([]bson.ObjectID(nil), roleIDs...)
	f.assignedBy = assignedBy
	return nil
}

func (f *fakeStaffRoleCreateRepository) ListActiveByRoleID(ctx context.Context, roleID bson.ObjectID) ([]authmodel.AdmStaffRole, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.rolesByRoleID, nil
}

func (f *fakeStaffRoleCreateRepository) ListActiveByStaffIDs(ctx context.Context, staffIDs []bson.ObjectID) ([]authmodel.AdmStaffRole, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.rolesByStaffIDs, nil
}

func (f *fakeStaffRoleCreateRepository) ReplaceByStaffID(ctx context.Context, staffID bson.ObjectID, roleIDs []bson.ObjectID, assignedBy bson.ObjectID) error {
	if f.err != nil {
		return f.err
	}
	f.replacedStaffID = staffID
	f.replacedRoleIDs = append([]bson.ObjectID(nil), roleIDs...)
	f.replacedAssignedBy = assignedBy
	return nil
}

func (f *fakeStaffRoleCreateRepository) RollbackCreateByStaffID(ctx context.Context, staffID bson.ObjectID) error {
	f.rollbackStaffID = staffID
	return nil
}

type fakeRoleCreateRepository struct {
	roles []authmodel.AdmRole
	err   error
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

func (f *fakeRoleCreateRepository) FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]authmodel.AdmRole, error) {
	if f.err != nil {
		return nil, f.err
	}
	roleByID := make(map[bson.ObjectID]authmodel.AdmRole, len(f.roles))
	for _, role := range f.roles {
		roleByID[role.ID] = role
	}
	result := make([]authmodel.AdmRole, 0, len(ids))
	for _, id := range ids {
		if role, ok := roleByID[id]; ok {
			result = append(result, role)
		}
	}
	return result, nil
}

func (f *fakeStaffRoleCreateRepository) String() string {
	return fmt.Sprintf("staff=%s roles=%v assigned_by=%s replaced_staff=%s replaced_roles=%v replaced_by=%s", f.createdStaffID.Hex(), f.createdRoleIDs, f.assignedBy.Hex(), f.replacedStaffID.Hex(), f.replacedRoleIDs, f.replacedAssignedBy.Hex())
}
