package auth

import (
	"fmt"

	commonmodel "house-manager/internal/model/common"
)

func (m *User) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("user is nil")
	}
	if commonmodel.IsBlank(m.Phone) {
		return fmt.Errorf("phone is required")
	}
	if !m.SourceChannel.ValidOptional() {
		return fmt.Errorf("sourceChannel is invalid")
	}
	return nil
}

func (m *UserAuth) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("user auth is nil")
	}
	if m.UserID.IsZero() {
		return fmt.Errorf("userID is required")
	}
	if !m.AuthProvider.Valid() {
		return fmt.Errorf("authProvider is invalid")
	}
	if commonmodel.IsBlank(m.OpenID) {
		return fmt.Errorf("openID is required")
	}
	return nil
}

func (m *UserProfileExt) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("user profile ext is nil")
	}
	if m.UserID.IsZero() {
		return fmt.Errorf("userID is required")
	}
	if m.BudgetMin < 0 || m.BudgetMax < 0 {
		return fmt.Errorf("budget must be non-negative")
	}
	return nil
}

func (m *Landlord) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("landlord is nil")
	}
	if commonmodel.IsBlank(m.Phone) {
		return fmt.Errorf("phone is required")
	}
	return nil
}

func (m *LandlordAuth) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("landlord auth is nil")
	}
	if m.LandlordID.IsZero() {
		return fmt.Errorf("landlordID is required")
	}
	if !m.AuthType.Valid() {
		return fmt.Errorf("authType is invalid")
	}
	if commonmodel.IsBlank(m.PasswordHash) {
		return fmt.Errorf("passwordHash is required")
	}
	return nil
}

func (m *LandlordProfile) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("landlord profile is nil")
	}
	if m.LandlordID.IsZero() {
		return fmt.Errorf("landlordID is required")
	}
	if commonmodel.IsBlank(m.LandlordName) {
		return fmt.Errorf("landlordName is required")
	}
	return nil
}

func (m *AdmStaff) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("adm staff is nil")
	}
	if commonmodel.IsBlank(m.Name) {
		return fmt.Errorf("name is required")
	}
	if commonmodel.IsBlank(m.Phone) {
		return fmt.Errorf("phone is required")
	}
	return nil
}

func (m *AdmStaffAuth) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("adm staff auth is nil")
	}
	if m.StaffID.IsZero() {
		return fmt.Errorf("staffID is required")
	}
	if !m.AuthType.Valid() {
		return fmt.Errorf("authType is invalid")
	}
	if commonmodel.IsBlank(m.PasswordHash) {
		return fmt.Errorf("passwordHash is required")
	}
	return nil
}

func (m *AdmRole) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("adm role is nil")
	}
	if commonmodel.IsBlank(m.RoleName) {
		return fmt.Errorf("roleName is required")
	}
	if commonmodel.IsBlank(m.RoleCode) {
		return fmt.Errorf("roleCode is required")
	}
	return nil
}

func (m *AdmPermission) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("adm permission is nil")
	}
	if commonmodel.IsBlank(m.PermissionName) {
		return fmt.Errorf("permissionName is required")
	}
	if commonmodel.IsBlank(m.PermissionCode) {
		return fmt.Errorf("permissionCode is required")
	}
	if commonmodel.IsBlank(m.Module) {
		return fmt.Errorf("module is required")
	}
	return nil
}

func (m *AdmStaffRole) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("adm staff role is nil")
	}
	if m.StaffID.IsZero() {
		return fmt.Errorf("staffID is required")
	}
	if m.RoleID.IsZero() {
		return fmt.Errorf("roleID is required")
	}
	return nil
}

func (m *AdmRolePermission) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("adm role permission is nil")
	}
	if m.RoleID.IsZero() {
		return fmt.Errorf("roleID is required")
	}
	if m.PermissionID.IsZero() {
		return fmt.Errorf("permissionID is required")
	}
	return nil
}

func (m *AdmLoginLog) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("adm login log is nil")
	}
	if m.StaffID.IsZero() {
		return fmt.Errorf("staffID is required")
	}
	if m.LoginAt <= 0 {
		return fmt.Errorf("loginAt is required")
	}
	if commonmodel.IsBlank(m.LoginIP) {
		return fmt.Errorf("loginIP is required")
	}
	if m.LoginResult != 1 && m.LoginResult != -1 {
		return fmt.Errorf("loginResult is invalid")
	}
	return nil
}
