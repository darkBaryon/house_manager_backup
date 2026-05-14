package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"house-manager/internal/model"
	repoaccount "house-manager/internal/repository/account"
	"house-manager/pkg/errcode"
	"house-manager/pkg/session"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	staffRepo          staffRepository
	roleRepo           roleRepository
	permissionRepo     permissionRepository
	staffRoleRepo      staffRoleRepository
	rolePermissionRepo rolePermissionRepository
	loginLogRepo       loginLogRepository
	sessionStore       *session.Store
	env                string
}

type staffRepository interface {
	FindActiveByPhone(ctx context.Context, phone string) (*model.AdmStaff, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*model.AdmStaff, error)
	TouchLastLogin(ctx context.Context, staffID bson.ObjectID, lastLoginAt int64, lastLoginIP string) error
}

type roleRepository interface {
	FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]model.AdmRole, error)
}

type permissionRepository interface {
	FindActiveByIDs(ctx context.Context, ids []bson.ObjectID) ([]model.AdmPermission, error)
}

type staffRoleRepository interface {
	ListActiveByStaffID(ctx context.Context, staffID bson.ObjectID) ([]model.AdmStaffRole, error)
}

type rolePermissionRepository interface {
	ListActiveByRoleIDs(ctx context.Context, roleIDs []bson.ObjectID) ([]model.AdmRolePermission, error)
}

type loginLogRepository interface {
	Create(ctx context.Context, log *model.AdmLoginLog) error
}

func NewService(
	staffRepo *repoaccount.StaffRepository,
	roleRepo *repoaccount.RoleRepository,
	permissionRepo *repoaccount.PermissionRepository,
	staffRoleRepo *repoaccount.StaffRoleRepository,
	rolePermissionRepo *repoaccount.RolePermissionRepository,
	loginLogRepo *repoaccount.LoginLogRepository,
	sessionStore *session.Store,
	env string,
) *Service {
	return newService(
		staffRepo,
		roleRepo,
		permissionRepo,
		staffRoleRepo,
		rolePermissionRepo,
		loginLogRepo,
		sessionStore,
		env,
	)
}

func newService(
	staffRepo staffRepository,
	roleRepo roleRepository,
	permissionRepo permissionRepository,
	staffRoleRepo staffRoleRepository,
	rolePermissionRepo rolePermissionRepository,
	loginLogRepo loginLogRepository,
	sessionStore *session.Store,
	env string,
) *Service {
	return &Service{
		staffRepo:          staffRepo,
		roleRepo:           roleRepo,
		permissionRepo:     permissionRepo,
		staffRoleRepo:      staffRoleRepo,
		rolePermissionRepo: rolePermissionRepo,
		loginLogRepo:       loginLogRepo,
		sessionStore:       sessionStore,
		env:                env,
	}
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	phone := strings.TrimSpace(input.Phone)
	if phone == "" {
		return nil, errcode.InvalidParam.WithError(fmt.Errorf("phone is required"))
	}
	if s.env != "local" {
		return nil, errcode.Forbidden.WithError(fmt.Errorf("phone-only publish login is local-only"))
	}
	if s.staffRepo == nil || s.sessionStore == nil {
		return nil, errcode.DatabaseError.WithError(fmt.Errorf("publish auth dependency is nil"))
	}

	staff, err := s.staffRepo.FindActiveByPhone(ctx, phone)
	if err != nil {
		return nil, errcode.DatabaseError.WithError(err)
	}
	if staff == nil {
		return nil, errcode.Unauthorized.WithError(fmt.Errorf("staff not found"))
	}

	authSession, err := s.buildSession(ctx, staff)
	if err != nil {
		return nil, err
	}
	token, err := s.sessionStore.CreatePrincipal(ctx, authSession.Principal)
	if err != nil {
		return nil, errcode.CacheError.WithError(err)
	}

	now := time.Now().Unix()
	if err := s.staffRepo.TouchLastLogin(ctx, staff.ID, now, input.LoginIP); err != nil {
		return nil, errcode.DatabaseError.WithError(err)
	}
	if s.loginLogRepo != nil {
		if err := s.loginLogRepo.Create(ctx, &model.AdmLoginLog{
			StaffID:     staff.ID,
			LoginAt:     now,
			LoginIP:     input.LoginIP,
			UserAgent:   input.UserAgent,
			LoginResult: 1,
		}); err != nil {
			return nil, errcode.DatabaseError.WithError(err)
		}
	}

	return &LoginResult{
		Token:       token,
		AuthSession: *authSession,
	}, nil
}

func (s *Service) Session(ctx context.Context, principal session.Principal) (*AuthSession, error) {
	if principal.Terminal != session.TerminalPublish || principal.PrincipalType != session.PrincipalTypeStaff {
		return nil, errcode.Unauthorized
	}
	staffID, err := bson.ObjectIDFromHex(principal.PrincipalID)
	if err != nil {
		return nil, errcode.Unauthorized.WithError(err)
	}
	staff, err := s.staffRepo.FindByID(ctx, staffID)
	if err != nil {
		return nil, errcode.DatabaseError.WithError(err)
	}
	if staff == nil || staff.Status != model.StatusActive {
		return nil, errcode.Unauthorized.WithError(fmt.Errorf("staff not found"))
	}
	return s.buildSession(ctx, staff)
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return errcode.Unauthorized
	}
	if err := s.sessionStore.Delete(ctx, token); err != nil {
		return errcode.CacheError.WithError(err)
	}
	return nil
}

func (s *Service) buildSession(ctx context.Context, staff *model.AdmStaff) (*AuthSession, error) {
	if staff == nil || staff.ID.IsZero() {
		return nil, errcode.DatabaseError.WithError(fmt.Errorf("staff is required"))
	}

	roleCodes, permissionCodes, err := s.loadRoleAndPermissionCodes(ctx, staff.ID)
	if err != nil {
		return nil, err
	}
	principal := session.Principal{
		PrincipalType:   session.PrincipalTypeStaff,
		PrincipalID:     staff.ID.Hex(),
		Terminal:        session.TerminalPublish,
		Phone:           staff.Phone,
		RoleCodes:       roleCodes,
		PermissionCodes: permissionCodes,
	}
	return &AuthSession{
		Principal: principal,
		Subject: Subject{
			ID:    staff.ID.Hex(),
			Name:  staff.Name,
			Phone: staff.Phone,
		},
	}, nil
}

func (s *Service) loadRoleAndPermissionCodes(ctx context.Context, staffID bson.ObjectID) ([]string, []string, error) {
	if s.staffRoleRepo == nil || s.roleRepo == nil || s.rolePermissionRepo == nil || s.permissionRepo == nil {
		return []string{}, []string{}, nil
	}
	staffRoles, err := s.staffRoleRepo.ListActiveByStaffID(ctx, staffID)
	if err != nil {
		return nil, nil, errcode.DatabaseError.WithError(err)
	}
	roleIDs := make([]bson.ObjectID, 0, len(staffRoles))
	for _, staffRole := range staffRoles {
		if !staffRole.RoleID.IsZero() {
			roleIDs = append(roleIDs, staffRole.RoleID)
		}
	}

	roles, err := s.roleRepo.FindActiveByIDs(ctx, roleIDs)
	if err != nil {
		return nil, nil, errcode.DatabaseError.WithError(err)
	}
	roleCodes := make([]string, 0, len(roles))
	activeRoleIDs := make([]bson.ObjectID, 0, len(roles))
	seenRoleCodes := map[string]struct{}{}
	for _, role := range roles {
		if role.RoleCode != "" {
			if _, ok := seenRoleCodes[role.RoleCode]; ok {
				continue
			}
			seenRoleCodes[role.RoleCode] = struct{}{}
			roleCodes = append(roleCodes, role.RoleCode)
		}
		if !role.ID.IsZero() {
			activeRoleIDs = append(activeRoleIDs, role.ID)
		}
	}

	rolePermissions, err := s.rolePermissionRepo.ListActiveByRoleIDs(ctx, activeRoleIDs)
	if err != nil {
		return nil, nil, errcode.DatabaseError.WithError(err)
	}
	permissionIDs := make([]bson.ObjectID, 0, len(rolePermissions))
	for _, rolePermission := range rolePermissions {
		if !rolePermission.PermissionID.IsZero() {
			permissionIDs = append(permissionIDs, rolePermission.PermissionID)
		}
	}

	permissions, err := s.permissionRepo.FindActiveByIDs(ctx, permissionIDs)
	if err != nil {
		return nil, nil, errcode.DatabaseError.WithError(err)
	}
	permissionCodes := make([]string, 0, len(permissions))
	seenPermissionCodes := map[string]struct{}{}
	for _, permission := range permissions {
		if permission.PermissionCode != "" {
			if _, ok := seenPermissionCodes[permission.PermissionCode]; ok {
				continue
			}
			seenPermissionCodes[permission.PermissionCode] = struct{}{}
			permissionCodes = append(permissionCodes, permission.PermissionCode)
		}
	}
	return roleCodes, permissionCodes, nil
}
