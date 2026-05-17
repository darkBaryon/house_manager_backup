package staff

import staffsvc "house-manager/internal/service/admin/staff"

type createRequest struct {
	Name          string   `json:"name"`
	Phone         string   `json:"phone"`
	Password      string   `json:"password"`
	Email         string   `json:"email"`
	Department    string   `json:"department"`
	JobTitle      string   `json:"job_title"`
	ContactQRCode string   `json:"contact_qr_code"`
	RoleIDs       []string `json:"role_ids"`
}

type listRequest struct {
	Keyword  string `json:"keyword"`
	Phone    string `json:"phone"`
	RoleID   string `json:"role_id"`
	Status   *int   `json:"status"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type detailRequest struct {
	StaffID string `json:"staff_id"`
}

type updateRequest struct {
	StaffID       string    `json:"staff_id"`
	Name          *string   `json:"name"`
	Email         *string   `json:"email"`
	Department    *string   `json:"department"`
	JobTitle      *string   `json:"job_title"`
	ContactQRCode *string   `json:"contact_qr_code"`
	Status        *int      `json:"status"`
	RoleIDs       *[]string `json:"role_ids"`
}

type disableRequest struct {
	StaffID string `json:"staff_id"`
}

type roleSummaryResponse struct {
	RoleID   string `json:"role_id"`
	RoleCode string `json:"role_code"`
	RoleName string `json:"role_name"`
}

type staffSummaryResponse struct {
	StaffID       string                `json:"staff_id"`
	Name          string                `json:"name"`
	Phone         string                `json:"phone"`
	Email         string                `json:"email"`
	Department    string                `json:"department"`
	JobTitle      string                `json:"job_title"`
	ContactQRCode string                `json:"contact_qr_code"`
	Roles         []roleSummaryResponse `json:"roles"`
	RoleNames     []string              `json:"role_names"`
	Status        int                   `json:"status"`
	CreatedAt     int64                 `json:"created_at"`
	UpdatedAt     int64                 `json:"updated_at"`
}

type createResponse struct {
	Staff staffSummaryResponse `json:"staff"`
}

type staffListItemResponse struct {
	StaffID       string                `json:"staff_id"`
	Name          string                `json:"name"`
	Phone         string                `json:"phone"`
	Email         string                `json:"email"`
	Department    string                `json:"department"`
	JobTitle      string                `json:"job_title"`
	ContactQRCode string                `json:"contact_qr_code"`
	Roles         []roleSummaryResponse `json:"roles"`
	RoleNames     []string              `json:"role_names"`
	Status        int                   `json:"status"`
	CreatedAt     int64                 `json:"created_at"`
	UpdatedAt     int64                 `json:"updated_at"`
	LastLoginAt   int64                 `json:"last_login_at"`
	LastLoginIP   string                `json:"last_login_ip"`
}

type listResponse struct {
	List     []staffListItemResponse `json:"list"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
	Total    int64                   `json:"total"`
}

type staffDetailResponse struct {
	StaffID           string                `json:"staff_id"`
	Name              string                `json:"name"`
	Phone             string                `json:"phone"`
	Email             string                `json:"email"`
	Department        string                `json:"department"`
	JobTitle          string                `json:"job_title"`
	ContactQRCode     string                `json:"contact_qr_code"`
	Roles             []roleSummaryResponse `json:"roles"`
	RoleNames         []string              `json:"role_names"`
	Status            int                   `json:"status"`
	CreatedByStaffID  string                `json:"created_by_staff_id"`
	CreatedAt         int64                 `json:"created_at"`
	UpdatedAt         int64                 `json:"updated_at"`
	PasswordUpdatedAt int64                 `json:"password_updated_at"`
	LastLoginAt       int64                 `json:"last_login_at"`
	LastLoginIP       string                `json:"last_login_ip"`
}

type detailResponse struct {
	Staff staffDetailResponse `json:"staff"`
}

type updateResponse struct {
	Staff staffDetailResponse `json:"staff"`
}

type disableResponse struct {
	Success bool `json:"success"`
}

func toCreateResponse(result *staffsvc.CreateResult) createResponse {
	return createResponse{
		Staff: toStaffSummaryResponse(result.Staff),
	}
}

func toListResponse(result *staffsvc.ListResult) listResponse {
	items := make([]staffListItemResponse, 0, len(result.List))
	for _, item := range result.List {
		items = append(items, toStaffListItemResponse(item))
	}
	if items == nil {
		items = []staffListItemResponse{}
	}
	return listResponse{
		List:     items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}
}

func toDetailResponse(result *staffsvc.DetailResult) detailResponse {
	return detailResponse{
		Staff: toStaffDetailResponse(result.Staff),
	}
}

func toUpdateResponse(result *staffsvc.UpdateResult) updateResponse {
	return updateResponse{
		Staff: toStaffDetailResponse(result.Staff),
	}
}

func toDisableResponse(result *staffsvc.DisableResult) disableResponse {
	return disableResponse{
		Success: result.Success,
	}
}

func toStaffSummaryResponse(staff staffsvc.StaffSummary) staffSummaryResponse {
	roles := make([]roleSummaryResponse, 0, len(staff.Roles))
	for _, role := range staff.Roles {
		roles = append(roles, roleSummaryResponse{
			RoleID:   role.RoleID,
			RoleCode: role.RoleCode,
			RoleName: role.RoleName,
		})
	}
	if roles == nil {
		roles = []roleSummaryResponse{}
	}
	return staffSummaryResponse{
		StaffID:       staff.StaffID,
		Name:          staff.Name,
		Phone:         staff.Phone,
		Email:         staff.Email,
		Department:    staff.Department,
		JobTitle:      staff.JobTitle,
		ContactQRCode: staff.ContactQRCode,
		Roles:         roles,
		RoleNames:     staff.RoleNames,
		Status:        staff.Status,
		CreatedAt:     staff.CreatedAt,
		UpdatedAt:     staff.UpdatedAt,
	}
}

func toStaffDetailResponse(staff staffsvc.StaffDetail) staffDetailResponse {
	summary := toStaffSummaryResponse(staff.StaffSummary)
	return staffDetailResponse{
		StaffID:           summary.StaffID,
		Name:              summary.Name,
		Phone:             summary.Phone,
		Email:             summary.Email,
		Department:        summary.Department,
		JobTitle:          summary.JobTitle,
		ContactQRCode:     summary.ContactQRCode,
		Roles:             summary.Roles,
		RoleNames:         summary.RoleNames,
		Status:            summary.Status,
		CreatedByStaffID:  staff.CreatedByStaffID,
		CreatedAt:         summary.CreatedAt,
		UpdatedAt:         summary.UpdatedAt,
		PasswordUpdatedAt: staff.PasswordUpdatedAt,
		LastLoginAt:       staff.LastLoginAt,
		LastLoginIP:       staff.LastLoginIP,
	}
}

func toStaffListItemResponse(item staffsvc.StaffListItem) staffListItemResponse {
	summary := toStaffSummaryResponse(item.StaffSummary)
	return staffListItemResponse{
		StaffID:       summary.StaffID,
		Name:          summary.Name,
		Phone:         summary.Phone,
		Email:         summary.Email,
		Department:    summary.Department,
		JobTitle:      summary.JobTitle,
		ContactQRCode: summary.ContactQRCode,
		Roles:         summary.Roles,
		RoleNames:     summary.RoleNames,
		Status:        summary.Status,
		CreatedAt:     summary.CreatedAt,
		UpdatedAt:     summary.UpdatedAt,
		LastLoginAt:   item.LastLoginAt,
		LastLoginIP:   item.LastLoginIP,
	}
}
