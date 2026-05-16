package auth

import (
	commonmodel "house-manager/internal/model/common"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	CollectionUser           = "hs_usr_user"
	CollectionUserAuth       = "hs_usr_auth"
	CollectionUserProfileExt = "hs_usr_profile_ext"

	CollectionLandlord        = "hs_lld_landlord"
	CollectionLandlordAuth    = "hs_lld_auth"
	CollectionLandlordProfile = "hs_lld_profile"

	CollectionAdmStaff          = "hs_adm_staff"
	CollectionAdmStaffAuth      = "hs_adm_staff_auth"
	CollectionAdmRole           = "hs_adm_role"
	CollectionAdmPermission     = "hs_adm_permission"
	CollectionAdmStaffRole      = "hs_adm_staff_role"
	CollectionAdmRolePermission = "hs_adm_role_permission"
	CollectionAdmLoginLog       = "hs_adm_login_log"
)

// User 对应 hs_usr_user，C 端用户主档。
type User struct {
	commonmodel.CommonFields `bson:",inline"`

	Nickname      string        `bson:"nickname" json:"nickname"`
	Avatar        string        `bson:"avatar" json:"avatar"`
	Phone         string        `bson:"phone" json:"phone"`
	City          string        `bson:"city" json:"city"`
	SourceChannel SourceChannel `bson:"source_channel" json:"sourceChannel"`
	LastActiveAt  int64         `bson:"last_active_at" json:"lastActiveAt"`
}

// UserAuth 对应 hs_usr_auth，微信身份认证绑定记录。
type UserAuth struct {
	commonmodel.CommonFields `bson:",inline"`

	UserID       bson.ObjectID `bson:"user_id" json:"userId"`
	AuthProvider AuthProvider  `bson:"auth_provider" json:"authProvider"`
	OpenID       string        `bson:"openid" json:"openid"`
	UnionID      string        `bson:"unionid,omitempty" json:"unionid"`
	LastLoginAt  int64         `bson:"last_login_at" json:"lastLoginAt"`
	LastLoginIP  string        `bson:"last_login_ip" json:"lastLoginIp"`
}

// UserProfileExt 对应 hs_usr_profile_ext，用户找房偏好扩展信息。
type UserProfileExt struct {
	commonmodel.CommonFields `bson:",inline"`

	UserID            bson.ObjectID `bson:"user_id" json:"userId"`
	BudgetMin         int           `bson:"budget_min" json:"budgetMin"`
	BudgetMax         int           `bson:"budget_max" json:"budgetMax"`
	PreferredAreas    []string      `bson:"preferred_areas" json:"preferredAreas"`
	PreferredRentMode string        `bson:"preferred_rent_mode" json:"preferredRentMode"`
	MoveInPlan        string        `bson:"move_in_plan" json:"moveInPlan"`
	Remark            string        `bson:"remark" json:"remark"`
}

// Landlord 对应 hs_lld_landlord，发房端房东主体。
type Landlord struct {
	commonmodel.CommonFields `bson:",inline"`

	Phone            string        `bson:"phone" json:"phone"`
	CreatedByStaffID bson.ObjectID `bson:"created_by_staff_id,omitempty" json:"createdByStaffId,omitempty"`
	UpdatedByStaffID bson.ObjectID `bson:"updated_by_staff_id,omitempty" json:"updatedByStaffId,omitempty"`
}

// LandlordAuth 对应 hs_lld_auth，房东认证信息。
type LandlordAuth struct {
	commonmodel.CommonFields `bson:",inline"`

	LandlordID        bson.ObjectID    `bson:"landlord_id" json:"landlordId"`
	AuthType          PasswordAuthType `bson:"auth_type" json:"authType"`
	PasswordHash      string           `bson:"password_hash" json:"-"`
	PasswordUpdatedAt int64            `bson:"password_updated_at" json:"passwordUpdatedAt"`
	LastLoginAt       int64            `bson:"last_login_at" json:"lastLoginAt"`
	LastLoginIP       string           `bson:"last_login_ip" json:"lastLoginIp"`
}

// LandlordProfile 对应 hs_lld_profile，房东业务资料。当前阶段暂不实现业务写入。
type LandlordProfile struct {
	commonmodel.CommonFields `bson:",inline"`

	LandlordID       bson.ObjectID `bson:"landlord_id" json:"landlordId"`
	LandlordName     string        `bson:"landlord_name" json:"landlordName"`
	CityCode         string        `bson:"city_code" json:"cityCode"`
	CityName         string        `bson:"city_name" json:"cityName"`
	DoorplateImages  []string      `bson:"doorplate_images" json:"doorplateImages"`
	AppearanceImages []string      `bson:"appearance_images" json:"appearanceImages"`
	CredentialImages []string      `bson:"credential_images" json:"credentialImages"`
	BrandName        string        `bson:"brand_name" json:"brandName"`
	BrandLogo        string        `bson:"brand_logo" json:"brandLogo"`
	BrandCover       string        `bson:"brand_cover" json:"brandCover"`
	BrandIntro       string        `bson:"brand_intro" json:"brandIntro"`
	CreatedByStaffID bson.ObjectID `bson:"created_by_staff_id,omitempty" json:"createdByStaffId,omitempty"`
	UpdatedByStaffID bson.ObjectID `bson:"updated_by_staff_id,omitempty" json:"updatedByStaffId,omitempty"`
}

// AdmStaff 对应 hs_adm_staff，后台员工主体。
type AdmStaff struct {
	commonmodel.CommonFields `bson:",inline"`

	Name             string        `bson:"name" json:"name"`
	Phone            string        `bson:"phone" json:"phone"`
	Email            string        `bson:"email" json:"email"`
	Department       string        `bson:"department" json:"department"`
	JobTitle         string        `bson:"job_title" json:"jobTitle"`
	ContactQRCode    string        `bson:"contact_qr_code" json:"contactQrCode"`
	CreatedByStaffID bson.ObjectID `bson:"created_by_staff_id,omitempty" json:"createdByStaffId,omitempty"`
}

// AdmStaffAuth 对应 hs_adm_staff_auth，后台员工认证信息。
type AdmStaffAuth struct {
	commonmodel.CommonFields `bson:",inline"`

	StaffID           bson.ObjectID    `bson:"staff_id" json:"staffId"`
	AuthType          PasswordAuthType `bson:"auth_type" json:"authType"`
	PasswordHash      string           `bson:"password_hash" json:"-"`
	PasswordUpdatedAt int64            `bson:"password_updated_at" json:"passwordUpdatedAt"`
	LastLoginAt       int64            `bson:"last_login_at" json:"lastLoginAt"`
	LastLoginIP       string           `bson:"last_login_ip" json:"lastLoginIp"`
}

// AdmRole 对应 hs_adm_role，角色定义。
type AdmRole struct {
	commonmodel.CommonFields `bson:",inline"`

	RoleName    string `bson:"role_name" json:"roleName"`
	RoleCode    string `bson:"role_code" json:"roleCode"`
	Description string `bson:"description" json:"description"`
	IsSystem    int    `bson:"is_system" json:"isSystem"`
}

// AdmPermission 对应 hs_adm_permission，权限点定义。
type AdmPermission struct {
	commonmodel.CommonFields `bson:",inline"`

	PermissionName string `bson:"permission_name" json:"permissionName"`
	PermissionCode string `bson:"permission_code" json:"permissionCode"`
	Module         string `bson:"module" json:"module"`
	Action         string `bson:"action" json:"action"`
	Description    string `bson:"description" json:"description"`
}

// AdmStaffRole 对应 hs_adm_staff_role，员工-角色关系。
type AdmStaffRole struct {
	commonmodel.CommonFields `bson:",inline"`

	StaffID    bson.ObjectID `bson:"staff_id" json:"staffId"`
	RoleID     bson.ObjectID `bson:"role_id" json:"roleId"`
	AssignedBy bson.ObjectID `bson:"assigned_by,omitempty" json:"assignedBy"`
	AssignedAt int64         `bson:"assigned_at" json:"assignedAt"`
}

// AdmRolePermission 对应 hs_adm_role_permission，角色-权限关系。
type AdmRolePermission struct {
	commonmodel.CommonFields `bson:",inline"`

	RoleID       bson.ObjectID `bson:"role_id" json:"roleId"`
	PermissionID bson.ObjectID `bson:"permission_id" json:"permissionId"`
	AssignedBy   bson.ObjectID `bson:"assigned_by,omitempty" json:"assignedBy"`
	AssignedAt   int64         `bson:"assigned_at" json:"assignedAt"`
}

// AdmLoginLog 对应 hs_adm_login_log，后台登录审计日志。
type AdmLoginLog struct {
	commonmodel.CommonFields `bson:",inline"`

	StaffID     bson.ObjectID `bson:"staff_id" json:"staffId"`
	LoginAt     int64         `bson:"login_at" json:"loginAt"`
	LoginIP     string        `bson:"login_ip" json:"loginIp"`
	UserAgent   string        `bson:"user_agent" json:"userAgent"`
	LoginResult int           `bson:"login_result" json:"loginResult"`
	Remark      string        `bson:"remark" json:"remark"`
}
