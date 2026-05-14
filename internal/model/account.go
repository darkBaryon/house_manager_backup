package model

import "go.mongodb.org/mongo-driver/v2/bson"

const (
	CollectionAdmStaff          = "hs_adm_staff"
	CollectionAdmRole           = "hs_adm_role"
	CollectionAdmPermission     = "hs_adm_permission"
	CollectionAdmStaffRole      = "hs_adm_staff_role"
	CollectionAdmRolePermission = "hs_adm_role_permission"
	CollectionAdmLoginLog       = "hs_adm_login_log"
)

// AdmStaff 对应 hs_adm_staff，后台员工档案。
type AdmStaff struct {
	CommonFields `bson:",inline"`

	Name          string        `bson:"name" json:"name"`
	Phone         string        `bson:"phone" json:"phone"`
	Email         string        `bson:"email" json:"email"`
	ContactQRCode string        `bson:"contact_qr_code" json:"contactQrCode"`
	LastLoginAt   int64         `bson:"last_login_at" json:"lastLoginAt"`
	LastLoginIP   string        `bson:"last_login_ip" json:"lastLoginIp"`
	CreatedBy     bson.ObjectID `bson:"created_by,omitempty" json:"createdBy"`
}

// AdmRole 对应 hs_adm_role，角色定义。
type AdmRole struct {
	CommonFields `bson:",inline"`

	RoleName    string `bson:"role_name" json:"roleName"`
	RoleCode    string `bson:"role_code" json:"roleCode"`
	Description string `bson:"description" json:"description"`
	IsSystem    int    `bson:"is_system" json:"isSystem"`
}

// AdmPermission 对应 hs_adm_permission，权限点定义。
type AdmPermission struct {
	CommonFields `bson:",inline"`

	PermissionName string `bson:"permission_name" json:"permissionName"`
	PermissionCode string `bson:"permission_code" json:"permissionCode"`
	Module         string `bson:"module" json:"module"`
	Action         string `bson:"action" json:"action"`
	Description    string `bson:"description" json:"description"`
}

// AdmStaffRole 对应 hs_adm_staff_role，员工-角色关系。
type AdmStaffRole struct {
	CommonFields `bson:",inline"`

	StaffID    bson.ObjectID `bson:"staff_id" json:"staffId"`
	RoleID     bson.ObjectID `bson:"role_id" json:"roleId"`
	AssignedBy bson.ObjectID `bson:"assigned_by,omitempty" json:"assignedBy"`
	AssignedAt int64         `bson:"assigned_at" json:"assignedAt"`
}

// AdmRolePermission 对应 hs_adm_role_permission，角色-权限关系。
type AdmRolePermission struct {
	CommonFields `bson:",inline"`

	RoleID       bson.ObjectID `bson:"role_id" json:"roleId"`
	PermissionID bson.ObjectID `bson:"permission_id" json:"permissionId"`
	AssignedBy   bson.ObjectID `bson:"assigned_by,omitempty" json:"assignedBy"`
	AssignedAt   int64         `bson:"assigned_at" json:"assignedAt"`
}

// AdmLoginLog 对应 hs_adm_login_log，后台登录审计日志。
type AdmLoginLog struct {
	CommonFields `bson:",inline"`

	StaffID     bson.ObjectID `bson:"staff_id" json:"staffId"`
	LoginAt     int64         `bson:"login_at" json:"loginAt"`
	LoginIP     string        `bson:"login_ip" json:"loginIp"`
	UserAgent   string        `bson:"user_agent" json:"userAgent"`
	LoginResult int           `bson:"login_result" json:"loginResult"`
	Remark      string        `bson:"remark" json:"remark"`
}
