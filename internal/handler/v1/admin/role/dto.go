package role

import rolesvc "house-manager/internal/service/admin/role"

type listRequest struct {
	Keyword  string `json:"keyword"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type detailRequest struct {
	RoleID string `json:"role_id"`
}

type createRequest struct {
	RoleName        string   `json:"role_name"`
	RoleCode        string   `json:"role_code"`
	Description     string   `json:"description"`
	PermissionCodes []string `json:"permission_codes"`
}

type updateRequest struct {
	RoleID          string    `json:"role_id"`
	RoleName        *string   `json:"role_name"`
	Description     *string   `json:"description"`
	PermissionCodes *[]string `json:"permission_codes"`
}

type permissionResponse struct {
	PermissionID   string `json:"permission_id"`
	PermissionCode string `json:"permission_code"`
	PermissionName string `json:"permission_name"`
	Module         string `json:"module"`
	Action         string `json:"action"`
}

type roleSummaryResponse struct {
	RoleID          string   `json:"role_id"`
	RoleName        string   `json:"role_name"`
	RoleCode        string   `json:"role_code"`
	Description     string   `json:"description"`
	PermissionCodes []string `json:"permission_codes"`
	PermissionNames []string `json:"permission_names"`
	IsSystem        int      `json:"is_system"`
	Status          int      `json:"status"`
	CreatedAt       int64    `json:"created_at"`
	UpdatedAt       int64    `json:"updated_at"`
}

type roleDetailResponse struct {
	RoleID          string               `json:"role_id"`
	RoleName        string               `json:"role_name"`
	RoleCode        string               `json:"role_code"`
	Description     string               `json:"description"`
	PermissionCodes []string             `json:"permission_codes"`
	PermissionNames []string             `json:"permission_names"`
	Permissions     []permissionResponse `json:"permissions"`
	IsSystem        int                  `json:"is_system"`
	Status          int                  `json:"status"`
	CreatedAt       int64                `json:"created_at"`
	UpdatedAt       int64                `json:"updated_at"`
}

type listItemResponse struct {
	roleSummaryResponse
}

type listResponse struct {
	List     []listItemResponse `json:"list"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int64              `json:"total"`
}

type detailResponse struct {
	Role roleDetailResponse `json:"role"`
}

type createResponse struct {
	Role roleDetailResponse `json:"role"`
}

type updateResponse struct {
	Role roleDetailResponse `json:"role"`
}

func toListResponse(result *rolesvc.ListResult) listResponse {
	items := make([]listItemResponse, 0, len(result.List))
	for _, item := range result.List {
		items = append(items, listItemResponse{roleSummaryResponse: toRoleSummaryResponse(item.RoleSummary)})
	}
	if items == nil {
		items = []listItemResponse{}
	}
	return listResponse{List: items, Page: result.Page, PageSize: result.PageSize, Total: result.Total}
}

func toDetailResponse(result *rolesvc.DetailResult) detailResponse {
	return detailResponse{Role: toRoleDetailResponse(result.Role)}
}

func toCreateResponse(result *rolesvc.CreateResult) createResponse {
	return createResponse{Role: toRoleDetailResponse(result.Role)}
}

func toUpdateResponse(result *rolesvc.UpdateResult) updateResponse {
	return updateResponse{Role: toRoleDetailResponse(result.Role)}
}

func toRoleSummaryResponse(role rolesvc.RoleSummary) roleSummaryResponse {
	return roleSummaryResponse{
		RoleID:          role.RoleID,
		RoleName:        role.RoleName,
		RoleCode:        role.RoleCode,
		Description:     role.Description,
		PermissionCodes: role.PermissionCodes,
		PermissionNames: role.PermissionNames,
		IsSystem:        role.IsSystem,
		Status:          role.Status,
		CreatedAt:       role.CreatedAt,
		UpdatedAt:       role.UpdatedAt,
	}
}

func toRoleDetailResponse(role rolesvc.RoleDetail) roleDetailResponse {
	permissions := make([]permissionResponse, 0, len(role.Permissions))
	for _, permission := range role.Permissions {
		permissions = append(permissions, permissionResponse{
			PermissionID:   permission.PermissionID,
			PermissionCode: permission.PermissionCode,
			PermissionName: permission.PermissionName,
			Module:         permission.Module,
			Action:         permission.Action,
		})
	}
	if permissions == nil {
		permissions = []permissionResponse{}
	}
	summary := toRoleSummaryResponse(role.RoleSummary)
	return roleDetailResponse{
		RoleID:          summary.RoleID,
		RoleName:        summary.RoleName,
		RoleCode:        summary.RoleCode,
		Description:     summary.Description,
		PermissionCodes: summary.PermissionCodes,
		PermissionNames: summary.PermissionNames,
		Permissions:     permissions,
		IsSystem:        summary.IsSystem,
		Status:          summary.Status,
		CreatedAt:       summary.CreatedAt,
		UpdatedAt:       summary.UpdatedAt,
	}
}
