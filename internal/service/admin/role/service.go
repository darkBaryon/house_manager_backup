package role

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	authmodel "house-manager/internal/model/auth"
	commonmodel "house-manager/internal/model/common"
	admrepo "house-manager/internal/repository/adm"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var roleCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)

type Service struct {
	roleRepo           roleRepository
	permissionRepo     permissionRepository
	rolePermissionRepo rolePermissionRepository
	staffRoleRepo      staffRoleRepository
	sessionStore       sessionInvalidator
}

type roleRepository interface {
	Create(ctx context.Context, role *authmodel.AdmRole) error
	FindActiveByID(ctx context.Context, id bson.ObjectID) (*authmodel.AdmRole, error)
	FindByCode(ctx context.Context, roleCode string) (*authmodel.AdmRole, error)
	List(ctx context.Context, input admrepo.RoleListFilter) ([]authmodel.AdmRole, int64, error)
	UpdateFields(ctx context.Context, id bson.ObjectID, fields bson.M) error
	RollbackCreate(ctx context.Context, id bson.ObjectID) error
}

type permissionRepository interface {
	FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]authmodel.AdmPermission, error)
	FindActiveByCodes(ctx context.Context, codes []string) ([]authmodel.AdmPermission, error)
}

type rolePermissionRepository interface {
	CreateMany(ctx context.Context, roleID bson.ObjectID, permissionIDs []bson.ObjectID, assignedBy bson.ObjectID) error
	ReplaceByRoleID(ctx context.Context, roleID bson.ObjectID, permissionIDs []bson.ObjectID, assignedBy bson.ObjectID) error
	ListActiveByRoleIDs(ctx context.Context, roleIDs []bson.ObjectID) ([]authmodel.AdmRolePermission, error)
	RollbackCreateByRoleID(ctx context.Context, roleID bson.ObjectID) error
}

type staffRoleRepository interface {
	ListActiveByRoleID(ctx context.Context, roleID bson.ObjectID) ([]authmodel.AdmStaffRole, error)
}

type sessionInvalidator interface {
	InvalidatePrincipal(ctx context.Context, principal session.Principal) error
}

func NewService(
	roleRepo *admrepo.RoleRepository,
	permissionRepo *admrepo.PermissionRepository,
	rolePermissionRepo *admrepo.RolePermissionRepository,
	staffRoleRepo *admrepo.StaffRoleRepository,
	sessionStore *session.Store,
) *Service {
	return newService(roleRepo, permissionRepo, rolePermissionRepo, staffRoleRepo, sessionStore)
}

func newService(
	roleRepo roleRepository,
	permissionRepo permissionRepository,
	rolePermissionRepo rolePermissionRepository,
	staffRoleRepo staffRoleRepository,
	sessionStores ...sessionInvalidator,
) *Service {
	var sessionStore sessionInvalidator
	if len(sessionStores) > 0 {
		sessionStore = sessionStores[0]
	}
	return &Service{
		roleRepo:           roleRepo,
		permissionRepo:     permissionRepo,
		rolePermissionRepo: rolePermissionRepo,
		staffRoleRepo:      staffRoleRepo,
		sessionStore:       sessionStore,
	}
}

func (s *Service) List(ctx context.Context, input ListInput) (*ListResult, error) {
	normalized := normalizeListInput(input)
	if err := s.requireRepos("获取角色列表失败，请稍后重试"); err != nil {
		return nil, err
	}

	roles, total, err := s.roleRepo.List(ctx, admrepo.RoleListFilter{
		Keyword: normalized.Keyword,
		Skip:    int64((normalized.Page - 1) * normalized.PageSize),
		Limit:   int64(normalized.PageSize),
	})
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "获取角色列表失败，请稍后重试", err)
	}
	if len(roles) == 0 {
		return &ListResult{List: []ListItem{}, Page: normalized.Page, PageSize: normalized.PageSize, Total: total}, nil
	}
	rolePermissions, permissions, err := s.loadPermissionsByRoles(ctx, collectRoleIDsFromRoles(roles), "获取角色列表失败，请稍后重试")
	if err != nil {
		return nil, err
	}
	items := make([]ListItem, 0, len(roles))
	for i := range roles {
		items = append(items, ListItem{RoleSummary: toRoleSummary(&roles[i], permissionsByRoleID(roles[i].ID, rolePermissions, permissions))})
	}
	return &ListResult{List: items, Page: normalized.Page, PageSize: normalized.PageSize, Total: total}, nil
}

func (s *Service) Detail(ctx context.Context, input DetailInput) (*DetailResult, error) {
	roleID, err := parseRoleID(input.RoleID)
	if err != nil {
		return nil, err
	}
	if err := s.requireRepos("获取角色详情失败，请稍后重试"); err != nil {
		return nil, err
	}
	detail, err := s.detailByID(ctx, roleID, "获取角色详情失败，请稍后重试")
	if err != nil {
		return nil, err
	}
	return &DetailResult{Role: *detail}, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*CreateResult, error) {
	normalized, err := normalizeCreateInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.requireRepos("创建角色失败，请稍后重试"); err != nil {
		return nil, err
	}
	operatorID, err := parseOperatorID(normalized.OperatorStaffID)
	if err != nil {
		return nil, err
	}

	existing, err := s.roleRepo.FindByCode(ctx, normalized.RoleCode)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "创建角色失败，请稍后重试", err)
	}
	if existing != nil {
		return nil, errcode.AlreadyExists.WithError(fmt.Errorf("角色编码已存在，请更换后重试"))
	}
	permissions, err := s.loadPermissionsByCodes(ctx, normalized.PermissionCodes, "创建角色失败，请稍后重试")
	if err != nil {
		return nil, err
	}

	role := &authmodel.AdmRole{
		RoleName:    normalized.RoleName,
		RoleCode:    normalized.RoleCode,
		Description: normalized.Description,
		IsSystem:    0,
	}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errcode.AlreadyExists.WithError(fmt.Errorf("角色编码已存在，请更换后重试"))
		}
		return nil, publicSystemError(errcode.DatabaseError.Code, "创建角色失败，请稍后重试", err)
	}
	if err := s.rolePermissionRepo.CreateMany(ctx, role.ID, collectPermissionIDs(permissions), operatorID); err != nil {
		return nil, s.rollbackCreate(ctx, role.ID, err)
	}
	return &CreateResult{Role: toRoleDetail(role, permissions)}, nil
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (*UpdateResult, error) {
	normalized, err := normalizeUpdateInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.requireRepos("更新角色失败，请稍后重试"); err != nil {
		return nil, err
	}
	roleID, err := parseRoleID(normalized.RoleID)
	if err != nil {
		return nil, err
	}
	operatorID, err := parseOperatorID(normalized.OperatorStaffID)
	if err != nil {
		return nil, err
	}
	role, err := s.roleRepo.FindActiveByID(ctx, roleID)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, "更新角色失败，请稍后重试", err)
	}
	if role == nil {
		return nil, errcode.NotFound.WithError(fmt.Errorf("角色不存在或已停用"))
	}

	var permissions []authmodel.AdmPermission
	if normalized.PermissionCodes != nil {
		permissions, err = s.loadPermissionsByCodes(ctx, *normalized.PermissionCodes, "更新角色失败，请稍后重试")
		if err != nil {
			return nil, err
		}
	}
	fields := updateFields(normalized)
	if len(fields) > 0 {
		if err := s.roleRepo.UpdateFields(ctx, roleID, fields); err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, errcode.NotFound.WithError(fmt.Errorf("角色不存在或已停用"))
			}
			return nil, publicSystemError(errcode.DatabaseError.Code, "更新角色失败，请稍后重试", err)
		}
	}
	if normalized.PermissionCodes != nil {
		if err := s.rolePermissionRepo.ReplaceByRoleID(ctx, roleID, collectPermissionIDs(permissions), operatorID); err != nil {
			return nil, publicSystemError(errcode.DatabaseError.Code, "更新角色失败，请稍后重试", err)
		}
		if err := s.invalidateStaffSessionsByRole(ctx, roleID, "更新角色失败，请稍后重试"); err != nil {
			return nil, err
		}
	}
	detail, err := s.detailByID(ctx, roleID, "更新角色失败，请稍后重试")
	if err != nil {
		return nil, err
	}
	return &UpdateResult{Role: *detail}, nil
}

func (s *Service) detailByID(ctx context.Context, roleID bson.ObjectID, systemMessage string) (*RoleDetail, error) {
	role, err := s.roleRepo.FindActiveByID(ctx, roleID)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, systemMessage, err)
	}
	if role == nil {
		return nil, errcode.NotFound.WithError(fmt.Errorf("角色不存在或已停用"))
	}
	rolePermissions, permissions, err := s.loadPermissionsByRoles(ctx, []bson.ObjectID{roleID}, systemMessage)
	if err != nil {
		return nil, err
	}
	return ptrRoleDetail(toRoleDetail(role, permissionsByRoleID(roleID, rolePermissions, permissions))), nil
}

func (s *Service) loadPermissionsByRoles(ctx context.Context, roleIDs []bson.ObjectID, systemMessage string) ([]authmodel.AdmRolePermission, []authmodel.AdmPermission, error) {
	rolePermissions, err := s.rolePermissionRepo.ListActiveByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, nil, publicSystemError(errcode.DatabaseError.Code, systemMessage, err)
	}
	permissions, err := s.permissionRepo.FindActiveByIDs(ctx, collectPermissionIDsFromRelations(rolePermissions))
	if err != nil {
		return nil, nil, publicSystemError(errcode.DatabaseError.Code, systemMessage, err)
	}
	return rolePermissions, permissions, nil
}

func (s *Service) loadPermissionsByCodes(ctx context.Context, codes []string, systemMessage string) ([]authmodel.AdmPermission, error) {
	permissions, err := s.permissionRepo.FindActiveByCodes(ctx, codes)
	if err != nil {
		return nil, publicSystemError(errcode.DatabaseError.Code, systemMessage, err)
	}
	if len(permissions) != len(codes) {
		return nil, errcode.InvalidParam.WithError(fmt.Errorf("所选权限不存在或已停用，请刷新后重试"))
	}
	byCode := make(map[string]authmodel.AdmPermission, len(permissions))
	for _, permission := range permissions {
		byCode[permission.PermissionCode] = permission
	}
	ordered := make([]authmodel.AdmPermission, 0, len(codes))
	for _, code := range codes {
		permission, ok := byCode[code]
		if !ok {
			return nil, errcode.InvalidParam.WithError(fmt.Errorf("所选权限不存在或已停用，请刷新后重试"))
		}
		ordered = append(ordered, permission)
	}
	return ordered, nil
}

func (s *Service) rollbackCreate(ctx context.Context, roleID bson.ObjectID, cause error) error {
	var rollbackErrs []error
	if s.rolePermissionRepo != nil {
		if err := s.rolePermissionRepo.RollbackCreateByRoleID(ctx, roleID); err != nil {
			rollbackErrs = append(rollbackErrs, err)
		}
	}
	if s.roleRepo != nil {
		if err := s.roleRepo.RollbackCreate(ctx, roleID); err != nil {
			rollbackErrs = append(rollbackErrs, err)
		}
	}
	if len(rollbackErrs) > 0 {
		return publicSystemError(errcode.DatabaseError.Code, "创建角色失败，请稍后重试", fmt.Errorf("create failed: %w; rollback failed: %w", cause, errors.Join(rollbackErrs...)))
	}
	return publicSystemError(errcode.DatabaseError.Code, "创建角色失败，请稍后重试", cause)
}

func (s *Service) invalidateStaffSessionsByRole(ctx context.Context, roleID bson.ObjectID, message string) error {
	if s.sessionStore == nil || s.staffRoleRepo == nil {
		return nil
	}
	staffRoles, err := s.staffRoleRepo.ListActiveByRoleID(ctx, roleID)
	if err != nil {
		return publicSystemError(errcode.DatabaseError.Code, message, err)
	}
	for _, staffRole := range staffRoles {
		if staffRole.StaffID.IsZero() {
			continue
		}
		if err := s.sessionStore.InvalidatePrincipal(ctx, session.Principal{
			PrincipalType: session.PrincipalTypeStaff,
			PrincipalID:   staffRole.StaffID.Hex(),
			Terminal:      session.TerminalAdmin,
		}); err != nil {
			return publicSystemError(errcode.CacheError.Code, message, err)
		}
	}
	return nil
}

func (s *Service) requireRepos(message string) error {
	if s.roleRepo == nil || s.permissionRepo == nil || s.rolePermissionRepo == nil {
		return publicSystemError(errcode.DatabaseError.Code, message, fmt.Errorf("后台角色服务未正确初始化"))
	}
	return nil
}

func normalizeListInput(input ListInput) ListInput {
	input.Keyword = strings.TrimSpace(input.Keyword)
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}
	return input
}

func normalizeCreateInput(input CreateInput) (CreateInput, error) {
	input.OperatorStaffID = strings.TrimSpace(input.OperatorStaffID)
	input.RoleName = strings.TrimSpace(input.RoleName)
	input.RoleCode = strings.TrimSpace(input.RoleCode)
	input.Description = strings.TrimSpace(input.Description)
	input.PermissionCodes = compactStrings(input.PermissionCodes)
	if input.RoleName == "" || input.RoleCode == "" || len(input.PermissionCodes) == 0 {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("请输入角色名称、角色编码和权限"))
	}
	if !roleCodePattern.MatchString(input.RoleCode) {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("角色编码格式不正确"))
	}
	return input, nil
}

func normalizeUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.OperatorStaffID = strings.TrimSpace(input.OperatorStaffID)
	input.RoleID = strings.TrimSpace(input.RoleID)
	if _, err := parseRoleID(input.RoleID); err != nil {
		return input, err
	}
	hasField := input.RoleName != nil || input.Description != nil || input.PermissionCodes != nil
	if !hasField {
		return input, errcode.InvalidParam.WithError(fmt.Errorf("请至少提交一个需要修改的字段"))
	}
	if input.RoleName != nil {
		roleName := strings.TrimSpace(*input.RoleName)
		if roleName == "" {
			return input, errcode.InvalidParam.WithError(fmt.Errorf("请输入角色名称"))
		}
		input.RoleName = &roleName
	}
	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		input.Description = &description
	}
	if input.PermissionCodes != nil {
		codes := compactStrings(*input.PermissionCodes)
		if len(codes) == 0 {
			return input, errcode.InvalidParam.WithError(fmt.Errorf("请至少选择一个权限"))
		}
		input.PermissionCodes = &codes
	}
	return input, nil
}

func parseRoleID(value string) (bson.ObjectID, error) {
	id, err := bson.ObjectIDFromHex(strings.TrimSpace(value))
	if err != nil || id.IsZero() {
		return bson.NilObjectID, errcode.InvalidParam.WithError(fmt.Errorf("角色参数不正确"))
	}
	return id, nil
}

func parseOperatorID(value string) (bson.ObjectID, error) {
	id, err := bson.ObjectIDFromHex(strings.TrimSpace(value))
	if err != nil || id.IsZero() {
		return bson.NilObjectID, errcode.Unauthorized.WithError(fmt.Errorf("当前登录状态无效，请重新登录"))
	}
	return id, nil
}

func updateFields(input UpdateInput) bson.M {
	fields := bson.M{}
	if input.RoleName != nil {
		fields["role_name"] = *input.RoleName
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	return fields
}

func toRoleDetail(role *authmodel.AdmRole, permissions []authmodel.AdmPermission) RoleDetail {
	return RoleDetail{
		RoleSummary: toRoleSummary(role, permissions),
		Permissions: toPermissionSummaries(permissions),
	}
}

func toRoleSummary(role *authmodel.AdmRole, permissions []authmodel.AdmPermission) RoleSummary {
	permissionSummaries := toPermissionSummaries(permissions)
	permissionCodes := make([]string, 0, len(permissionSummaries))
	permissionNames := make([]string, 0, len(permissionSummaries))
	for _, permission := range permissionSummaries {
		permissionCodes = append(permissionCodes, permission.PermissionCode)
		permissionNames = append(permissionNames, permission.PermissionName)
	}
	status := role.Status
	if status == commonmodel.StatusUnspecified {
		status = commonmodel.StatusActive
	}
	return RoleSummary{
		RoleID:          role.ID.Hex(),
		RoleName:        role.RoleName,
		RoleCode:        role.RoleCode,
		Description:     role.Description,
		PermissionCodes: permissionCodes,
		PermissionNames: permissionNames,
		IsSystem:        role.IsSystem,
		Status:          status,
		CreatedAt:       role.CreatedAt,
		UpdatedAt:       role.UpdatedAt,
	}
}

func toPermissionSummaries(permissions []authmodel.AdmPermission) []PermissionSummary {
	items := make([]PermissionSummary, 0, len(permissions))
	for _, permission := range permissions {
		items = append(items, PermissionSummary{
			PermissionID:   permission.ID.Hex(),
			PermissionCode: permission.PermissionCode,
			PermissionName: permission.PermissionName,
			Module:         permission.Module,
			Action:         permission.Action,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].PermissionCode < items[j].PermissionCode
	})
	if items == nil {
		items = []PermissionSummary{}
	}
	return items
}

func permissionsByRoleID(roleID bson.ObjectID, relations []authmodel.AdmRolePermission, permissions []authmodel.AdmPermission) []authmodel.AdmPermission {
	permissionByID := make(map[bson.ObjectID]authmodel.AdmPermission, len(permissions))
	for _, permission := range permissions {
		permissionByID[permission.ID] = permission
	}
	items := make([]authmodel.AdmPermission, 0)
	for _, relation := range relations {
		if relation.RoleID != roleID {
			continue
		}
		permission, ok := permissionByID[relation.PermissionID]
		if !ok {
			continue
		}
		items = append(items, permission)
	}
	return items
}

func collectRoleIDsFromRoles(roles []authmodel.AdmRole) []bson.ObjectID {
	items := make([]bson.ObjectID, 0, len(roles))
	for _, role := range roles {
		if !role.ID.IsZero() {
			items = append(items, role.ID)
		}
	}
	return items
}

func collectPermissionIDs(permissions []authmodel.AdmPermission) []bson.ObjectID {
	items := make([]bson.ObjectID, 0, len(permissions))
	for _, permission := range permissions {
		if !permission.ID.IsZero() {
			items = append(items, permission.ID)
		}
	}
	return items
}

func collectPermissionIDsFromRelations(relations []authmodel.AdmRolePermission) []bson.ObjectID {
	items := make([]bson.ObjectID, 0, len(relations))
	seen := make(map[bson.ObjectID]struct{}, len(relations))
	for _, relation := range relations {
		if relation.PermissionID.IsZero() {
			continue
		}
		if _, ok := seen[relation.PermissionID]; ok {
			continue
		}
		seen[relation.PermissionID] = struct{}{}
		items = append(items, relation.PermissionID)
	}
	return items
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

func ptrRoleDetail(value RoleDetail) *RoleDetail {
	return &value
}

func publicSystemError(code int, message string, cause error) *errcode.Error {
	return errcode.New(code, message).WithError(cause)
}
