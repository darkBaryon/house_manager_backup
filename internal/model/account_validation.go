package model

import "fmt"

func (m *AdmStaff) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("adm staff is nil")
	}
	if isBlank(m.Name) {
		return fmt.Errorf("name is required")
	}
	if isBlank(m.Phone) {
		return fmt.Errorf("phone is required")
	}
	return nil
}

func (m *AdmRole) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("adm role is nil")
	}
	if isBlank(m.RoleName) {
		return fmt.Errorf("roleName is required")
	}
	if isBlank(m.RoleCode) {
		return fmt.Errorf("roleCode is required")
	}
	return nil
}

func (m *AdmPermission) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("adm permission is nil")
	}
	if isBlank(m.PermissionName) {
		return fmt.Errorf("permissionName is required")
	}
	if isBlank(m.PermissionCode) {
		return fmt.Errorf("permissionCode is required")
	}
	if isBlank(m.Module) {
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
	if isBlank(m.LoginIP) {
		return fmt.Errorf("loginIP is required")
	}
	if m.LoginResult != 1 && m.LoginResult != -1 {
		return fmt.Errorf("loginResult is invalid")
	}
	return nil
}
