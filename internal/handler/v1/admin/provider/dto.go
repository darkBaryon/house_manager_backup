package provider

import providersvc "house-manager/internal/service/admin/provider"

type createRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type listRequest struct {
	Phone    string `json:"phone"`
	Status   *int   `json:"status"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type detailRequest struct {
	ProviderID string `json:"provider_id"`
}

type updateRequest struct {
	ProviderID string `json:"provider_id"`
	Phone      string `json:"phone"`
}

type disableRequest struct {
	ProviderID string `json:"provider_id"`
}

type providerSummaryResponse struct {
	ProviderID        string `json:"provider_id"`
	Phone             string `json:"phone"`
	Status            int    `json:"status"`
	CreatedByStaffID  string `json:"created_by_staff_id"`
	UpdatedByStaffID  string `json:"updated_by_staff_id"`
	CreatedAt         int64  `json:"created_at"`
	UpdatedAt         int64  `json:"updated_at"`
	PasswordUpdatedAt int64  `json:"password_updated_at"`
	LastLoginAt       int64  `json:"last_login_at"`
	LastLoginIP       string `json:"last_login_ip"`
}

type createResponse struct {
	Provider providerSummaryResponse `json:"provider"`
}

type listItemResponse struct {
	ProviderID        string `json:"provider_id"`
	Phone             string `json:"phone"`
	Status            int    `json:"status"`
	CreatedByStaffID  string `json:"created_by_staff_id"`
	UpdatedByStaffID  string `json:"updated_by_staff_id"`
	CreatedAt         int64  `json:"created_at"`
	UpdatedAt         int64  `json:"updated_at"`
	PasswordUpdatedAt int64  `json:"password_updated_at"`
	LastLoginAt       int64  `json:"last_login_at"`
	LastLoginIP       string `json:"last_login_ip"`
}

type listResponse struct {
	List     []listItemResponse `json:"list"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int64              `json:"total"`
}

type detailResponse struct {
	Provider providerSummaryResponse `json:"provider"`
}

type updateResponse struct {
	Provider providerSummaryResponse `json:"provider"`
}

type disableResponse struct {
	Success bool `json:"success"`
}

func toCreateResponse(result *providersvc.CreateResult) createResponse {
	return createResponse{Provider: toProviderSummaryResponse(result.Provider)}
}

func toListResponse(result *providersvc.ListResult) listResponse {
	items := make([]listItemResponse, 0, len(result.List))
	for _, item := range result.List {
		items = append(items, toListItemResponse(item))
	}
	if items == nil {
		items = []listItemResponse{}
	}
	return listResponse{
		List:     items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}
}

func toDetailResponse(result *providersvc.DetailResult) detailResponse {
	return detailResponse{Provider: toProviderSummaryResponse(result.Provider)}
}

func toUpdateResponse(result *providersvc.UpdateResult) updateResponse {
	return updateResponse{Provider: toProviderSummaryResponse(result.Provider)}
}

func toDisableResponse(result *providersvc.DisableResult) disableResponse {
	return disableResponse{Success: result.Success}
}

func toProviderSummaryResponse(provider providersvc.ProviderSummary) providerSummaryResponse {
	return providerSummaryResponse{
		ProviderID:        provider.ProviderID,
		Phone:             provider.Phone,
		Status:            provider.Status,
		CreatedByStaffID:  provider.CreatedByStaffID,
		UpdatedByStaffID:  provider.UpdatedByStaffID,
		CreatedAt:         provider.CreatedAt,
		UpdatedAt:         provider.UpdatedAt,
		PasswordUpdatedAt: provider.PasswordUpdatedAt,
		LastLoginAt:       provider.LastLoginAt,
		LastLoginIP:       provider.LastLoginIP,
	}
}

func toListItemResponse(item providersvc.ListItem) listItemResponse {
	summary := toProviderSummaryResponse(item.ProviderSummary)
	return listItemResponse{
		ProviderID:        summary.ProviderID,
		Phone:             summary.Phone,
		Status:            summary.Status,
		CreatedByStaffID:  summary.CreatedByStaffID,
		UpdatedByStaffID:  summary.UpdatedByStaffID,
		CreatedAt:         summary.CreatedAt,
		UpdatedAt:         summary.UpdatedAt,
		PasswordUpdatedAt: summary.PasswordUpdatedAt,
		LastLoginAt:       summary.LastLoginAt,
		LastLoginIP:       summary.LastLoginIP,
	}
}
