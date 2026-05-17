package auth

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	authmodel "house-manager/internal/model/auth"
	admrepo "house-manager/internal/repository/adm"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	staffRepo          staffRepository
	staffAuthRepo      staffAuthRepository
	staffRoleRepo      staffRoleRepository
	roleRepo           roleRepository
	rolePermissionRepo rolePermissionRepository
	permissionRepo     permissionRepository
	loginLogRepo       loginLogRepository
	sessionStore       *session.Store
}

type staffRepository interface {
	FindActiveByPhone(ctx context.Context, phone string) (*authmodel.AdmStaff, error)
	FindActiveByID(ctx context.Context, id bson.ObjectID) (*authmodel.AdmStaff, error)
}

type staffAuthRepository interface {
	FindActivePasswordByStaffID(ctx context.Context, staffID bson.ObjectID) (*authmodel.AdmStaffAuth, error)
	TouchLastLogin(ctx context.Context, authID bson.ObjectID, loginIP string) error
}

type staffRoleRepository interface {
	ListActiveByStaffID(ctx context.Context, staffID bson.ObjectID) ([]authmodel.AdmStaffRole, error)
}

type roleRepository interface {
	FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]authmodel.AdmRole, error)
}

type rolePermissionRepository interface {
	ListActiveByRoleIDs(ctx context.Context, roleIDs []bson.ObjectID) ([]authmodel.AdmRolePermission, error)
}

type permissionRepository interface {
	FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]authmodel.AdmPermission, error)
}

type loginLogRepository interface {
	Create(ctx context.Context, logEntry *authmodel.AdmLoginLog) error
}

func NewService(
	staffRepo *admrepo.StaffRepository,
	staffAuthRepo *admrepo.StaffAuthRepository,
	staffRoleRepo *admrepo.StaffRoleRepository,
	roleRepo *admrepo.RoleRepository,
	rolePermissionRepo *admrepo.RolePermissionRepository,
	permissionRepo *admrepo.PermissionRepository,
	loginLogRepo *admrepo.LoginLogRepository,
	sessionStore *session.Store,
) *Service {
	return newService(
		staffRepo,
		staffAuthRepo,
		staffRoleRepo,
		roleRepo,
		rolePermissionRepo,
		permissionRepo,
		loginLogRepo,
		sessionStore,
	)
}

func newService(
	staffRepo staffRepository,
	staffAuthRepo staffAuthRepository,
	staffRoleRepo staffRoleRepository,
	roleRepo roleRepository,
	rolePermissionRepo rolePermissionRepository,
	permissionRepo permissionRepository,
	loginLogRepo loginLogRepository,
	sessionStore *session.Store,
) *Service {
	return &Service{
		staffRepo:          staffRepo,
		staffAuthRepo:      staffAuthRepo,
		staffRoleRepo:      staffRoleRepo,
		roleRepo:           roleRepo,
		rolePermissionRepo: rolePermissionRepo,
		permissionRepo:     permissionRepo,
		loginLogRepo:       loginLogRepo,
		sessionStore:       sessionStore,
	}
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	phone := strings.TrimSpace(input.Phone)
	password := strings.TrimSpace(input.Password)
	logAuthInfo(ctx, "admin.auth.login.start", "phone", maskPhone(phone))

	if phone == "" || password == "" {
		err := errcode.InvalidParam.WithError(fmt.Errorf("请输入手机号和密码"))
		logAuthResult(ctx, "admin.auth.login.success", "admin.auth.login.failed", err, "phone", maskPhone(phone))
		return nil, err
	}
	if s.staffRepo == nil || s.staffAuthRepo == nil || s.staffRoleRepo == nil || s.roleRepo == nil || s.rolePermissionRepo == nil || s.permissionRepo == nil || s.sessionStore == nil {
		err := publicSystemError(errcode.DatabaseError.Code, "登录失败，请稍后重试", fmt.Errorf("后台登录服务未正确初始化"))
		logAuthResult(ctx, "admin.auth.login.success", "admin.auth.login.failed", err, "phone", maskPhone(phone))
		return nil, err
	}

	staff, err := s.staffRepo.FindActiveByPhone(ctx, phone)
	if err != nil {
		dbErr := publicSystemError(errcode.DatabaseError.Code, "登录失败，请稍后重试", err)
		logAuthResult(ctx, "admin.auth.login.success", "admin.auth.login.failed", dbErr, "phone", maskPhone(phone))
		return nil, dbErr
	}
	if staff == nil {
		authErr := errcode.Unauthorized.WithError(fmt.Errorf("手机号或密码错误，请重新输入"))
		logAuthResult(ctx, "admin.auth.login.success", "admin.auth.login.failed", authErr, "phone", maskPhone(phone))
		return nil, authErr
	}

	authRecord, err := s.staffAuthRepo.FindActivePasswordByStaffID(ctx, staff.ID)
	if err != nil {
		dbErr := publicSystemError(errcode.DatabaseError.Code, "登录失败，请稍后重试", err)
		logAuthResult(ctx, "admin.auth.login.success", "admin.auth.login.failed", dbErr, "staff_id", staff.ID.Hex())
		return nil, dbErr
	}
	if authRecord == nil || bcrypt.CompareHashAndPassword([]byte(authRecord.PasswordHash), []byte(password)) != nil {
		authErr := errcode.Unauthorized.WithError(fmt.Errorf("手机号或密码错误，请重新输入"))
		logAuthResult(ctx, "admin.auth.login.success", "admin.auth.login.failed", authErr, "staff_id", staff.ID.Hex())
		return nil, authErr
	}

	authSession, err := s.buildSession(ctx, staff, "登录失败，请稍后重试")
	if err != nil {
		logAuthResult(ctx, "admin.auth.login.success", "admin.auth.login.failed", err, "staff_id", staff.ID.Hex())
		return nil, err
	}

	token, err := s.sessionStore.CreatePrincipal(ctx, authSession.Principal)
	if err != nil {
		cacheErr := publicSystemError(errcode.CacheError.Code, "登录失败，请稍后重试", err)
		logAuthResult(ctx, "admin.auth.login.success", "admin.auth.login.failed", cacheErr, "staff_id", staff.ID.Hex())
		return nil, cacheErr
	}

	if err := s.staffAuthRepo.TouchLastLogin(ctx, authRecord.ID, input.LoginIP); err != nil {
		logAuthResult(ctx, "admin.auth.login.success", "admin.auth.login.failed", errcode.DatabaseError.WithError(err), "staff_id", staff.ID.Hex(), "step", "touch_last_login")
	}
	if err := s.writeLoginLog(ctx, staff.ID, input.LoginIP, input.UserAgent); err != nil {
		logAuthResult(ctx, "admin.auth.login.success", "admin.auth.login.failed", errcode.DatabaseError.WithError(err), "staff_id", staff.ID.Hex(), "step", "write_login_log")
	}

	result := &LoginResult{
		Token:       token,
		AuthSession: *authSession,
	}
	logAuthInfo(ctx, "admin.auth.login.success", "staff_id", staff.ID.Hex())
	return result, nil
}

func (s *Service) Session(ctx context.Context, principal session.Principal) (*AuthSession, error) {
	logAuthInfo(ctx, "admin.auth.session.start")
	if principal.Terminal != session.TerminalAdmin || principal.PrincipalType != session.PrincipalTypeStaff {
		err := errcode.Unauthorized.WithError(fmt.Errorf("当前登录状态无效，请重新登录"))
		logAuthResult(ctx, "admin.auth.session.success", "admin.auth.session.failed", err, "principal_id", principal.PrincipalID)
		return nil, err
	}
	if s.staffRepo == nil || s.staffRoleRepo == nil || s.roleRepo == nil || s.rolePermissionRepo == nil || s.permissionRepo == nil {
		err := publicSystemError(errcode.DatabaseError.Code, "获取登录状态失败，请稍后重试", fmt.Errorf("后台登录状态服务未正确初始化"))
		logAuthResult(ctx, "admin.auth.session.success", "admin.auth.session.failed", err, "principal_id", principal.PrincipalID)
		return nil, err
	}
	staffID, err := bson.ObjectIDFromHex(principal.PrincipalID)
	if err != nil {
		authErr := errcode.Unauthorized.WithError(fmt.Errorf("当前登录状态无效，请重新登录"))
		logAuthResult(ctx, "admin.auth.session.success", "admin.auth.session.failed", authErr, "principal_id", principal.PrincipalID)
		return nil, authErr
	}
	staff, err := s.staffRepo.FindActiveByID(ctx, staffID)
	if err != nil {
		dbErr := publicSystemError(errcode.DatabaseError.Code, "获取登录状态失败，请稍后重试", err)
		logAuthResult(ctx, "admin.auth.session.success", "admin.auth.session.failed", dbErr, "staff_id", staffID.Hex())
		return nil, dbErr
	}
	if staff == nil {
		authErr := errcode.Unauthorized.WithError(fmt.Errorf("当前登录状态已失效，请重新登录"))
		logAuthResult(ctx, "admin.auth.session.success", "admin.auth.session.failed", authErr, "staff_id", staffID.Hex())
		return nil, authErr
	}
	authSession, err := s.buildSession(ctx, staff, "获取登录状态失败，请稍后重试")
	logAuthResult(ctx, "admin.auth.session.success", "admin.auth.session.failed", err, "staff_id", staffID.Hex())
	return authSession, err
}

func (s *Service) Logout(ctx context.Context, token string) error {
	logAuthInfo(ctx, "admin.auth.logout.start")
	if strings.TrimSpace(token) == "" {
		err := errcode.Unauthorized.WithError(fmt.Errorf("未登录或登录已过期，请重新登录"))
		logAuthResult(ctx, "admin.auth.logout.success", "admin.auth.logout.failed", err)
		return err
	}
	if s.sessionStore == nil {
		err := publicSystemError(errcode.CacheError.Code, "退出登录失败，请稍后重试", fmt.Errorf("后台退出登录服务未正确初始化"))
		logAuthResult(ctx, "admin.auth.logout.success", "admin.auth.logout.failed", err)
		return err
	}
	if err := s.sessionStore.Delete(ctx, token); err != nil {
		cacheErr := publicSystemError(errcode.CacheError.Code, "退出登录失败，请稍后重试", err)
		logAuthResult(ctx, "admin.auth.logout.success", "admin.auth.logout.failed", cacheErr)
		return cacheErr
	}
	logAuthInfo(ctx, "admin.auth.logout.success")
	return nil
}

func (s *Service) buildSession(ctx context.Context, staff *authmodel.AdmStaff, publicMessage string) (*AuthSession, error) {
	if staff == nil || staff.ID.IsZero() {
		return nil, publicSystemError(errcode.DatabaseError.Code, publicMessage, fmt.Errorf("员工主体数据无效"))
	}

	roleCodes, permissionCodes, err := s.loadRoleAndPermissionCodes(ctx, staff.ID, publicMessage)
	if err != nil {
		return nil, err
	}

	principal := session.Principal{
		PrincipalType:   session.PrincipalTypeStaff,
		PrincipalID:     staff.ID.Hex(),
		Terminal:        session.TerminalAdmin,
		Phone:           staff.Phone,
		RoleCodes:       roleCodes,
		PermissionCodes: permissionCodes,
	}

	return &AuthSession{
		Principal:       principal,
		StaffProfile:    toStaffProfile(staff),
		RoleCodes:       roleCodes,
		PermissionCodes: permissionCodes,
	}, nil
}

func (s *Service) loadRoleAndPermissionCodes(ctx context.Context, staffID bson.ObjectID, publicMessage string) ([]string, []string, error) {
	staffRoles, err := s.staffRoleRepo.ListActiveByStaffID(ctx, staffID)
	if err != nil {
		return nil, nil, publicSystemError(errcode.DatabaseError.Code, publicMessage, err)
	}
	roleIDs := make([]bson.ObjectID, 0, len(staffRoles))
	for _, item := range staffRoles {
		roleIDs = append(roleIDs, item.RoleID)
	}

	roles, err := s.roleRepo.FindActiveByIDs(ctx, roleIDs)
	if err != nil {
		return nil, nil, publicSystemError(errcode.DatabaseError.Code, publicMessage, err)
	}
	roleCodes := collectSortedRoleCodes(roles)

	rolePermissions, err := s.rolePermissionRepo.ListActiveByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, nil, publicSystemError(errcode.DatabaseError.Code, publicMessage, err)
	}
	permissionIDs := make([]bson.ObjectID, 0, len(rolePermissions))
	for _, item := range rolePermissions {
		permissionIDs = append(permissionIDs, item.PermissionID)
	}

	permissions, err := s.permissionRepo.FindActiveByIDs(ctx, permissionIDs)
	if err != nil {
		return nil, nil, publicSystemError(errcode.DatabaseError.Code, publicMessage, err)
	}
	permissionCodes := collectSortedPermissionCodes(permissions)
	return roleCodes, permissionCodes, nil
}

func (s *Service) writeLoginLog(ctx context.Context, staffID bson.ObjectID, loginIP string, userAgent string) error {
	if s.loginLogRepo == nil {
		return nil
	}
	return s.loginLogRepo.Create(ctx, &authmodel.AdmLoginLog{
		StaffID:     staffID,
		LoginAt:     time.Now().Unix(),
		LoginIP:     loginIP,
		UserAgent:   userAgent,
		LoginResult: 1,
		Remark:      "",
	})
}

func toStaffProfile(staff *authmodel.AdmStaff) StaffProfile {
	return StaffProfile{
		StaffID:       staff.ID.Hex(),
		Name:          staff.Name,
		Phone:         staff.Phone,
		Email:         staff.Email,
		Department:    staff.Department,
		JobTitle:      staff.JobTitle,
		ContactQRCode: staff.ContactQRCode,
	}
}

func collectSortedRoleCodes(items []authmodel.AdmRole) []string {
	codes := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		code := strings.TrimSpace(item.RoleCode)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	slices.Sort(codes)
	return codes
}

func collectSortedPermissionCodes(items []authmodel.AdmPermission) []string {
	codes := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		code := strings.TrimSpace(item.PermissionCode)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	slices.Sort(codes)
	return codes
}

func publicSystemError(code int, message string, cause error) *errcode.Error {
	return errcode.New(code, message).WithError(cause)
}
