package staff

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	admrepo "house-manager/internal/repository/adm"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = bcrypt.DefaultCost

var mainlandPhonePattern = regexp.MustCompile(`^1[3-9][0-9]{9}$`)

type Service struct {
	staffRepo     staffRepository
	staffAuthRepo staffAuthRepository
	staffRoleRepo staffRoleRepository
	roleRepo      roleRepository
	sessionStore  sessionInvalidator
}

type staffRepository interface {
	Create(ctx context.Context, staff *authmodel.AdmStaff) error
	FindByPhone(ctx context.Context, phone string) (*authmodel.AdmStaff, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*authmodel.AdmStaff, error)
	List(ctx context.Context, input admrepo.StaffListFilter) ([]authmodel.AdmStaff, int64, error)
	UpdateFields(ctx context.Context, id bson.ObjectID, fields bson.M) error
	RollbackCreate(ctx context.Context, id bson.ObjectID) error
}

type staffAuthRepository interface {
	CreatePasswordAuth(ctx context.Context, authRecord *authmodel.AdmStaffAuth) error
	FindActivePasswordByStaffIDs(ctx context.Context, staffIDs []bson.ObjectID) ([]authmodel.AdmStaffAuth, error)
	RollbackCreateByStaffID(ctx context.Context, staffID bson.ObjectID) error
}

type staffRoleRepository interface {
	CreateMany(ctx context.Context, staffID bson.ObjectID, roleIDs []bson.ObjectID, assignedBy bson.ObjectID) error
	ReplaceByStaffID(ctx context.Context, staffID bson.ObjectID, roleIDs []bson.ObjectID, assignedBy bson.ObjectID) error
	ListActiveByRoleID(ctx context.Context, roleID bson.ObjectID) ([]authmodel.AdmStaffRole, error)
	ListActiveByStaffIDs(ctx context.Context, staffIDs []bson.ObjectID) ([]authmodel.AdmStaffRole, error)
	RollbackCreateByStaffID(ctx context.Context, staffID bson.ObjectID) error
}

type roleRepository interface {
	FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]authmodel.AdmRole, error)
}

type sessionInvalidator interface {
	InvalidatePrincipal(ctx context.Context, principal session.Principal) error
}

func NewService(
	staffRepo *admrepo.StaffRepository,
	staffAuthRepo *admrepo.StaffAuthRepository,
	staffRoleRepo *admrepo.StaffRoleRepository,
	roleRepo *admrepo.RoleRepository,
	sessionStore *session.Store,
) *Service {
	return newService(staffRepo, staffAuthRepo, staffRoleRepo, roleRepo, sessionStore)
}

func newService(
	staffRepo staffRepository,
	staffAuthRepo staffAuthRepository,
	staffRoleRepo staffRoleRepository,
	roleRepo roleRepository,
	sessionStores ...sessionInvalidator,
) *Service {
	var sessionStore sessionInvalidator
	if len(sessionStores) > 0 {
		sessionStore = sessionStores[0]
	}
	return &Service{
		staffRepo:     staffRepo,
		staffAuthRepo: staffAuthRepo,
		staffRoleRepo: staffRoleRepo,
		roleRepo:      roleRepo,
		sessionStore:  sessionStore,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*CreateResult, error) {
	normalized, err := normalizeCreateInput(input)
	if err != nil {
		return nil, err
	}
	if s.staffRepo == nil || s.staffAuthRepo == nil || s.staffRoleRepo == nil || s.roleRepo == nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "创建员工失败，请稍后重试", fmt.Errorf("后台员工服务未正确初始化"))
	}

	operatorID, err := bson.ObjectIDFromHex(normalized.OperatorStaffID)
	if err != nil {
		return nil, errcode.Unauthorized.WithError(fmt.Errorf("当前登录状态无效，请重新登录"))
	}

	existing, err := s.staffRepo.FindByPhone(ctx, normalized.Phone)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "创建员工失败，请稍后重试", err)
	}
	if existing != nil {
		return nil, errcode.AlreadyExists.WithError(fmt.Errorf("员工手机号已存在，请更换后重试"))
	}

	roleIDs, err := parseRoleIDs(normalized.RoleIDs)
	if err != nil {
		return nil, err
	}
	roles, err := s.loadRoles(ctx, roleIDs, "创建员工失败，请稍后重试")
	if err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(normalized.Password), bcryptCost)
	if err != nil {
		return nil, publicSystemError(errcode.InternalError.Code, "创建员工失败，请稍后重试", err)
	}

	now := time.Now().Unix()
	staffRecord := &authmodel.AdmStaff{
		Name:             normalized.Name,
		Phone:            normalized.Phone,
		Email:            normalized.Email,
		Department:       normalized.Department,
		JobTitle:         normalized.JobTitle,
		ContactQRCode:    normalized.ContactQRCode,
		CreatedByStaffID: operatorID,
	}
	if err := s.staffRepo.Create(ctx, staffRecord); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errcode.AlreadyExists.WithError(fmt.Errorf("员工手机号已存在，请更换后重试"))
		}
		return nil, publicSystemError(errcode.DatabaseError.Code, "创建员工失败，请稍后重试", err)
	}

	authRecord := &authmodel.AdmStaffAuth{
		StaffID:           staffRecord.ID,
		AuthType:          authmodel.PasswordAuthTypePassword,
		PasswordHash:      string(passwordHash),
		PasswordUpdatedAt: now,
	}
	if err := s.staffAuthRepo.CreatePasswordAuth(ctx, authRecord); err != nil {
		return nil, s.rollbackCreate(ctx, staffRecord.ID, err)
	}

	if err := s.staffRoleRepo.CreateMany(ctx, staffRecord.ID, roleIDs, operatorID); err != nil {
		return nil, s.rollbackCreate(ctx, staffRecord.ID, err)
	}

	return &CreateResult{
		Staff: toStaffSummary(staffRecord, roles),
	}, nil
}

func (s *Service) rollbackCreate(ctx context.Context, staffID bson.ObjectID, cause error) error {
	var rollbackErrs []error
	if s.staffRoleRepo != nil {
		if err := s.staffRoleRepo.RollbackCreateByStaffID(ctx, staffID); err != nil {
			rollbackErrs = append(rollbackErrs, err)
		}
	}
	if s.staffAuthRepo != nil {
		if err := s.staffAuthRepo.RollbackCreateByStaffID(ctx, staffID); err != nil {
			rollbackErrs = append(rollbackErrs, err)
		}
	}
	if s.staffRepo != nil {
		if err := s.staffRepo.RollbackCreate(ctx, staffID); err != nil {
			rollbackErrs = append(rollbackErrs, err)
		}
	}
	if len(rollbackErrs) > 0 {
		return publicSystemError(errcode.DatabaseError.Code, "创建员工失败，请稍后重试", fmt.Errorf("create failed: %w; rollback failed: %w", cause, errors.Join(rollbackErrs...)))
	}
	return publicSystemError(errcode.DatabaseError.Code, "创建员工失败，请稍后重试", cause)
}

func (s *Service) List(ctx context.Context, input ListInput) (*ListResult, error) {
	normalized, err := normalizeListInput(input)
	if err != nil {
		return nil, err
	}
	if s.staffRepo == nil || s.staffAuthRepo == nil || s.staffRoleRepo == nil || s.roleRepo == nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "获取员工列表失败，请稍后重试", fmt.Errorf("后台员工服务未正确初始化"))
	}

	var staffIDs []bson.ObjectID
	if normalized.RoleID != "" {
		roleID, err := bson.ObjectIDFromHex(normalized.RoleID)
		if err != nil {
			return nil, errcode.InvalidParam.WithError(fmt.Errorf("角色参数不正确"))
		}
		staffRoles, err := s.staffRoleRepo.ListActiveByRoleID(ctx, roleID)
		if err != nil {
			return nil, publicSystemError(errcode.DatabaseError.Code, "获取员工列表失败，请稍后重试", err)
		}
		staffIDs = collectStaffIDs(staffRoles)
		if len(staffIDs) == 0 {
			return emptyListResult(normalized), nil
		}
	}

	staffItems, total, err := s.staffRepo.List(ctx, admrepo.StaffListFilter{
		Keyword:  normalized.Keyword,
		Phone:    normalized.Phone,
		Status:   *normalized.Status,
		StaffIDs: staffIDs,
		Skip:     int64((normalized.Page - 1) * normalized.PageSize),
		Limit:    int64(normalized.PageSize),
	})
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "获取员工列表失败，请稍后重试", err)
	}
	if len(staffItems) == 0 {
		return &ListResult{List: []StaffListItem{}, Page: normalized.Page, PageSize: normalized.PageSize, Total: total}, nil
	}

	pageStaffIDs := collectStaffIDsFromStaff(staffItems)
	staffRoles, err := s.staffRoleRepo.ListActiveByStaffIDs(ctx, pageStaffIDs)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "获取员工列表失败，请稍后重试", err)
	}
	roles, err := s.roleRepo.FindActiveByIDs(ctx, collectRoleIDs(staffRoles))
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "获取员工列表失败，请稍后重试", err)
	}
	authRecords, err := s.staffAuthRepo.FindActivePasswordByStaffIDs(ctx, pageStaffIDs)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "获取员工列表失败，请稍后重试", err)
	}

	return &ListResult{
		List:     buildStaffListItems(staffItems, staffRoles, roles, authRecords),
		Page:     normalized.Page,
		PageSize: normalized.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) Detail(ctx context.Context, input DetailInput) (*DetailResult, error) {
	staffID, err := parseStaffID(input.StaffID)
	if err != nil {
		return nil, err
	}
	if s.staffRepo == nil || s.staffAuthRepo == nil || s.staffRoleRepo == nil || s.roleRepo == nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "获取员工详情失败，请稍后重试", fmt.Errorf("后台员工服务未正确初始化"))
	}

	detail, err := s.detailByID(ctx, staffID, "获取员工详情失败，请稍后重试")
	if err != nil {
		return nil, err
	}
	return &DetailResult{Staff: *detail}, nil
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (*UpdateResult, error) {
	normalized, err := normalizeUpdateInput(input)
	if err != nil {
		return nil, err
	}
	if s.staffRepo == nil || s.staffAuthRepo == nil || s.staffRoleRepo == nil || s.roleRepo == nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "更新员工失败，请稍后重试", fmt.Errorf("后台员工服务未正确初始化"))
	}

	staffID, err := parseStaffID(normalized.StaffID)
	if err != nil {
		return nil, err
	}
	staff, err := s.staffRepo.FindByID(ctx, staffID)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "更新员工失败，请稍后重试", err)
	}
	if staff == nil {
		return nil, errcode.NotFound.WithError(fmt.Errorf("员工不存在或已删除"))
	}

	var roleIDs []bson.ObjectID
	if normalized.RoleIDs != nil {
		roleIDs, err = parseRoleIDs(*normalized.RoleIDs)
		if err != nil {
			return nil, err
		}
		if _, err := s.loadRoles(ctx, roleIDs, "更新员工失败，请稍后重试"); err != nil {
			return nil, err
		}
	}

	fields := updateFields(normalized)
	if len(fields) > 0 {
		if err := s.staffRepo.UpdateFields(ctx, staffID, fields); err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, errcode.NotFound.WithError(fmt.Errorf("员工不存在或已删除"))
			}
			return nil, publicSystemError(errcode.DatabaseError.Code, "更新员工失败，请稍后重试", err)
		}
	}

	if normalized.RoleIDs != nil {
		operatorID, err := bson.ObjectIDFromHex(normalized.OperatorStaffID)
		if err != nil {
			return nil, errcode.Unauthorized.WithError(fmt.Errorf("当前登录状态无效，请重新登录"))
		}
		if err := s.staffRoleRepo.ReplaceByStaffID(ctx, staffID, roleIDs, operatorID); err != nil {
			return nil, publicSystemError(errcode.DatabaseError.Code, "更新员工失败，请稍后重试", err)
		}
	}
	if normalized.RoleIDs != nil || (normalized.Status != nil && *normalized.Status == commonmodel.StatusDeleted) {
		if err := s.invalidateStaffSession(ctx, staffID, "更新员工失败，请稍后重试"); err != nil {
			return nil, err
		}
	}

	detail, err := s.detailByID(ctx, staffID, "更新员工失败，请稍后重试")
	if err != nil {
		return nil, err
	}
	return &UpdateResult{Staff: *detail}, nil
}

func (s *Service) Disable(ctx context.Context, input DisableInput) (*DisableResult, error) {
	staffID, err := parseStaffID(input.StaffID)
	if err != nil {
		return nil, err
	}
	operatorID, err := parseStaffID(input.OperatorStaffID)
	if err != nil {
		return nil, errcode.Unauthorized.WithError(fmt.Errorf("当前登录状态无效，请重新登录"))
	}
	if staffID == operatorID {
		return nil, errcode.InvalidParam.WithError(fmt.Errorf("不能禁用当前登录账号"))
	}
	if s.staffRepo == nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "禁用员工失败，请稍后重试", fmt.Errorf("后台员工服务未正确初始化"))
	}

	staff, err := s.staffRepo.FindByID(ctx, staffID)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "禁用员工失败，请稍后重试", err)
	}
	if staff == nil {
		return nil, errcode.NotFound.WithError(fmt.Errorf("员工不存在或已删除"))
	}
	if err := s.staffRepo.UpdateFields(ctx, staffID, bson.M{"status": commonmodel.StatusDeleted}); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errcode.NotFound.WithError(fmt.Errorf("员工不存在或已删除"))
		}
		return nil, publicSystemError(errcode.DatabaseError.Code, "禁用员工失败，请稍后重试", err)
	}
	if err := s.invalidateStaffSession(ctx, staffID, "禁用员工失败，请稍后重试"); err != nil {
		return nil, err
	}
	return &DisableResult{Success: true}, nil
}

func (s *Service) invalidateStaffSession(ctx context.Context, staffID bson.ObjectID, message string) error {
	if s.sessionStore == nil {
		return nil
	}
	if err := s.sessionStore.InvalidatePrincipal(ctx, session.Principal{
		PrincipalType: session.PrincipalTypeStaff,
		PrincipalID:   staffID.Hex(),
		Terminal:      session.TerminalAdmin,
	}); err != nil {
		return publicSystemError(errcode.CacheError.Code, message, err)
	}
	return nil
}

func (s *Service) detailByID(ctx context.Context, staffID bson.ObjectID, systemMessage string) (*StaffDetail, error) {
	staff, err := s.staffRepo.FindByID(ctx, staffID)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, systemMessage, err)
	}
	if staff == nil {
		return nil, errcode.NotFound.WithError(fmt.Errorf("员工不存在或已删除"))
	}

	staffRoles, err := s.staffRoleRepo.ListActiveByStaffIDs(ctx, []bson.ObjectID{staffID})
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, systemMessage, err)
	}
	roles, err := s.roleRepo.FindActiveByIDs(ctx, collectRoleIDs(staffRoles))
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, systemMessage, err)
	}
	authRecords, err := s.staffAuthRepo.FindActivePasswordByStaffIDs(ctx, []bson.ObjectID{staffID})
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, systemMessage, err)
	}

	detail := toStaffDetail(staff, roles)
	if len(authRecords) > 0 {
		detail.PasswordUpdatedAt = authRecords[0].PasswordUpdatedAt
		detail.LastLoginAt = authRecords[0].LastLoginAt
		detail.LastLoginIP = authRecords[0].LastLoginIP
	}
	return &detail, nil
}

func normalizeCreateInput(input CreateInput) (CreateInput, error) {
	input.OperatorStaffID = strings.TrimSpace(input.OperatorStaffID)
	input.Name = strings.TrimSpace(input.Name)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Password = strings.TrimSpace(input.Password)
	input.Email = strings.TrimSpace(input.Email)
	input.Department = strings.TrimSpace(input.Department)
	input.JobTitle = strings.TrimSpace(input.JobTitle)
	input.ContactQRCode = strings.TrimSpace(input.ContactQRCode)
	if input.Name == "" || input.Phone == "" || input.Password == "" {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("请输入员工姓名、手机号和初始密码"))
	}
	if !mainlandPhonePattern.MatchString(input.Phone) {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("请输入正确的员工手机号"))
	}
	input.RoleIDs = compactStrings(input.RoleIDs)
	return input, nil
}

func normalizeListInput(input ListInput) (ListInput, error) {
	input.Keyword = strings.TrimSpace(input.Keyword)
	input.Phone = strings.TrimSpace(input.Phone)
	input.RoleID = strings.TrimSpace(input.RoleID)
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}
	if input.Status == nil {
		status := commonmodel.StatusActive
		input.Status = &status
	}
	if *input.Status != commonmodel.StatusActive && *input.Status != commonmodel.StatusDeleted {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("员工状态参数不正确"))
	}
	if input.RoleID != "" {
		if _, err := bson.ObjectIDFromHex(input.RoleID); err != nil {
			return input, errcode.InvalidParam.WithError(fmt.Errorf("角色参数不正确"))
		}
	}
	return input, nil
}

func normalizeUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.OperatorStaffID = strings.TrimSpace(input.OperatorStaffID)
	input.StaffID = strings.TrimSpace(input.StaffID)
	if _, err := parseStaffID(input.StaffID); err != nil {
		return input, err
	}
	hasField := input.Name != nil ||
		input.Email != nil ||
		input.Department != nil ||
		input.JobTitle != nil ||
		input.ContactQRCode != nil ||
		input.Status != nil ||
		input.RoleIDs != nil
	if !hasField {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("请至少提交一个需要修改的字段"))
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return input, errcode.InvalidParam.WithError(fmt.Errorf("请输入员工姓名"))
		}
		input.Name = &name
	}
	if input.Email != nil {
		email := strings.TrimSpace(*input.Email)
		input.Email = &email
	}
	if input.Department != nil {
		department := strings.TrimSpace(*input.Department)
		input.Department = &department
	}
	if input.JobTitle != nil {
		jobTitle := strings.TrimSpace(*input.JobTitle)
		input.JobTitle = &jobTitle
	}
	if input.ContactQRCode != nil {
		contactQRCode := strings.TrimSpace(*input.ContactQRCode)
		input.ContactQRCode = &contactQRCode
	}
	if input.Status != nil && *input.Status != commonmodel.StatusActive && *input.Status != commonmodel.StatusDeleted {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("员工状态参数不正确"))
	}
	if input.RoleIDs != nil {
		roleIDs := compactStrings(*input.RoleIDs)
		input.RoleIDs = &roleIDs
	}
	return input, nil
}

func parseRoleIDs(values []string) ([]bson.ObjectID, error) {
	if len(values) == 0 {
		return []bson.ObjectID{}, nil
	}
	result := make([]bson.ObjectID, 0, len(values))
	for _, value := range values {
		id, err := bson.ObjectIDFromHex(value)
		if err != nil || id.IsZero() {
			return nil, errcode.InvalidParam.WithError(fmt.Errorf("角色参数不正确"))
		}
		result = append(result, id)
	}
	return result, nil
}

func parseStaffID(value string) (bson.ObjectID, error) {
	id, err := bson.ObjectIDFromHex(strings.TrimSpace(value))
	if err != nil || id.IsZero() {
		return bson.NilObjectID, errcode.InvalidParam.WithError(fmt.Errorf("员工参数不正确"))
	}
	return id, nil
}

func (s *Service) loadRoles(ctx context.Context, roleIDs []bson.ObjectID, systemMessage string) ([]authmodel.AdmRole, error) {
	if len(roleIDs) == 0 {
		return []authmodel.AdmRole{}, nil
	}
	roles, err := s.roleRepo.FindActiveByIDs(ctx, roleIDs)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, systemMessage, err)
	}
	if len(roles) != len(roleIDs) {
		return nil, errcode.InvalidParam.WithError(fmt.Errorf("所选角色不存在或已停用，请刷新后重试"))
	}
	roleByID := make(map[bson.ObjectID]authmodel.AdmRole, len(roles))
	for _, role := range roles {
		roleByID[role.ID] = role
	}
	ordered := make([]authmodel.AdmRole, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		role, ok := roleByID[roleID]
		if !ok {
			return nil, errcode.InvalidParam.WithError(fmt.Errorf("所选角色不存在或已停用，请刷新后重试"))
		}
		ordered = append(ordered, role)
	}
	return ordered, nil
}

func updateFields(input UpdateInput) bson.M {
	fields := bson.M{}
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Email != nil {
		fields["email"] = *input.Email
	}
	if input.Department != nil {
		fields["department"] = *input.Department
	}
	if input.JobTitle != nil {
		fields["job_title"] = *input.JobTitle
	}
	if input.ContactQRCode != nil {
		fields["contact_qr_code"] = *input.ContactQRCode
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	return fields
}

func toStaffSummary(staff *authmodel.AdmStaff, roles []authmodel.AdmRole) StaffSummary {
	roleSummaries, roleNames := toRoleSummaries(roles)
	status := staff.Status
	if status == commonmodel.StatusUnspecified {
		status = commonmodel.StatusActive
	}
	return StaffSummary{
		StaffID:       staff.ID.Hex(),
		Name:          staff.Name,
		Phone:         staff.Phone,
		Email:         staff.Email,
		Department:    staff.Department,
		JobTitle:      staff.JobTitle,
		ContactQRCode: staff.ContactQRCode,
		Roles:         roleSummaries,
		RoleNames:     roleNames,
		Status:        status,
		CreatedAt:     staff.CreatedAt,
		UpdatedAt:     staff.UpdatedAt,
	}
}

func toStaffDetail(staff *authmodel.AdmStaff, roles []authmodel.AdmRole) StaffDetail {
	detail := StaffDetail{
		StaffSummary: toStaffSummary(staff, roles),
	}
	if !staff.CreatedByStaffID.IsZero() {
		detail.CreatedByStaffID = staff.CreatedByStaffID.Hex()
	}
	return detail
}

func toRoleSummaries(roles []authmodel.AdmRole) ([]RoleSummary, []string) {
	roleSummaries := make([]RoleSummary, 0, len(roles))
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleSummaries = append(roleSummaries, RoleSummary{
			RoleID:   role.ID.Hex(),
			RoleCode: role.RoleCode,
			RoleName: role.RoleName,
		})
		roleNames = append(roleNames, role.RoleName)
	}
	sort.SliceStable(roleSummaries, func(i, j int) bool {
		return roleSummaries[i].RoleCode < roleSummaries[j].RoleCode
	})
	sort.Strings(roleNames)
	if roleSummaries == nil {
		roleSummaries = []RoleSummary{}
	}
	if roleNames == nil {
		roleNames = []string{}
	}
	return roleSummaries, roleNames
}

func emptyListResult(input ListInput) *ListResult {
	return &ListResult{
		List:     []StaffListItem{},
		Page:     input.Page,
		PageSize: input.PageSize,
		Total:    0,
	}
}

func collectStaffIDs(items []authmodel.AdmStaffRole) []bson.ObjectID {
	result := make([]bson.ObjectID, 0, len(items))
	seen := make(map[bson.ObjectID]struct{}, len(items))
	for _, item := range items {
		if item.StaffID.IsZero() {
			continue
		}
		if _, ok := seen[item.StaffID]; ok {
			continue
		}
		seen[item.StaffID] = struct{}{}
		result = append(result, item.StaffID)
	}
	return result
}

func collectStaffIDsFromStaff(items []authmodel.AdmStaff) []bson.ObjectID {
	result := make([]bson.ObjectID, 0, len(items))
	for _, item := range items {
		if !item.ID.IsZero() {
			result = append(result, item.ID)
		}
	}
	return result
}

func collectRoleIDs(items []authmodel.AdmStaffRole) []bson.ObjectID {
	result := make([]bson.ObjectID, 0, len(items))
	seen := make(map[bson.ObjectID]struct{}, len(items))
	for _, item := range items {
		if item.RoleID.IsZero() {
			continue
		}
		if _, ok := seen[item.RoleID]; ok {
			continue
		}
		seen[item.RoleID] = struct{}{}
		result = append(result, item.RoleID)
	}
	return result
}

func buildStaffListItems(
	staffItems []authmodel.AdmStaff,
	staffRoles []authmodel.AdmStaffRole,
	roles []authmodel.AdmRole,
	authRecords []authmodel.AdmStaffAuth,
) []StaffListItem {
	roleByID := make(map[bson.ObjectID]authmodel.AdmRole, len(roles))
	for _, role := range roles {
		roleByID[role.ID] = role
	}
	rolesByStaffID := make(map[bson.ObjectID][]authmodel.AdmRole, len(staffRoles))
	for _, staffRole := range staffRoles {
		role, ok := roleByID[staffRole.RoleID]
		if !ok {
			continue
		}
		rolesByStaffID[staffRole.StaffID] = append(rolesByStaffID[staffRole.StaffID], role)
	}
	authByStaffID := make(map[bson.ObjectID]authmodel.AdmStaffAuth, len(authRecords))
	for _, authRecord := range authRecords {
		authByStaffID[authRecord.StaffID] = authRecord
	}

	result := make([]StaffListItem, 0, len(staffItems))
	for _, staff := range staffItems {
		summary := toStaffSummary(&staff, rolesByStaffID[staff.ID])
		item := StaffListItem{StaffSummary: summary}
		if authRecord, ok := authByStaffID[staff.ID]; ok {
			item.LastLoginAt = authRecord.LastLoginAt
			item.LastLoginIP = authRecord.LastLoginIP
		}
		result = append(result, item)
	}
	return result
}

func compactStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || slices.Contains(result, value) {
			continue
		}
		result = append(result, value)
	}
	return result
}

func publicSystemError(code int, message string, cause error) *errcode.Error {
	return errcode.New(code, message).WithError(cause)
}
