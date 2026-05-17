package house

import housesvc "house-manager/internal/service/admin/house"

type listRequest struct {
	ProviderID    string `json:"provider_id"`
	AssetMode     string `json:"asset_mode"`
	City          string `json:"city"`
	District      string `json:"district"`
	RoomStatus    *int   `json:"room_status"`
	ListingStatus *int   `json:"listing_status"`
	AuditStatus   *int   `json:"audit_status"`
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
}

type detailRequest struct {
	ListingID string `json:"listing_id"`
}

type listItemResponse struct {
	ListingID     string `json:"listing_id"`
	AssetMode     string `json:"asset_mode"`
	ProviderID    string `json:"provider_id"`
	ProviderPhone string `json:"provider_phone"`
	Title         string `json:"title"`
	City          string `json:"city"`
	District      string `json:"district"`
	BizArea       string `json:"biz_area"`
	CommunityName string `json:"community_name"`
	BuildingName  string `json:"building_name"`
	RoomNo        string `json:"room_no"`
	Price         int    `json:"price"`
	PriceText     string `json:"price_text"`
	LayoutText    string `json:"layout_text"`
	AreaSize      int    `json:"area_size"`
	RoomStatus    int    `json:"room_status"`
	ListingStatus int    `json:"listing_status"`
	AuditStatus   int    `json:"audit_status"`
	IsOnline      int    `json:"is_online"`
	UpdatedAt     int64  `json:"updated_at"`
}

type listResponse struct {
	List     []listItemResponse `json:"list"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int64              `json:"total"`
}

type detailResponse struct {
	House houseResponse `json:"house"`
}

type houseResponse struct {
	ListingID         string `json:"listing_id"`
	SourceType        string `json:"source_type"`
	SourceID          string `json:"source_id"`
	AssetMode         string `json:"asset_mode"`
	ProviderID        string `json:"provider_id"`
	ProviderPhone     string `json:"provider_phone"`
	LandlordName      string `json:"landlord_name"`
	RootType          string `json:"root_type"`
	RootID            string `json:"root_id"`
	ProjectID         string `json:"project_id"`
	ProjectName       string `json:"project_name"`
	BuildingID        string `json:"building_id"`
	BuildingName      string `json:"building_name"`
	RoomTypeID        string `json:"room_type_id"`
	RoomTypeName      string `json:"room_type_name"`
	DecentralizedID   string `json:"decentralized_id"`
	CommunityName     string `json:"community_name"`
	RentMode          string `json:"rent_mode"`
	City              string `json:"city"`
	District          string `json:"district"`
	BizArea           string `json:"biz_area"`
	AddressText       string `json:"address_text"`
	RoomNo            string `json:"room_no"`
	Title             string `json:"title"`
	Price             int    `json:"price"`
	PriceText         string `json:"price_text"`
	LayoutText        string `json:"layout_text"`
	AreaSize          int    `json:"area_size"`
	RoomStatus        int    `json:"room_status"`
	ListingStatus     int    `json:"listing_status"`
	AuditStatus       int    `json:"audit_status"`
	IsOnline          int    `json:"is_online"`
	LatestAuditTaskID string `json:"latest_audit_task_id"`
	LatestSubmittedAt int64  `json:"latest_submitted_at"`
	LatestReviewedAt  int64  `json:"latest_reviewed_at"`
	ReviewerStaffID   string `json:"reviewer_staff_id"`
	CreatedAt         int64  `json:"created_at"`
	UpdatedAt         int64  `json:"updated_at"`
}

func toListResponse(result *housesvc.ListResult) listResponse {
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

func toDetailResponse(result *housesvc.DetailResult) detailResponse {
	return detailResponse{House: toHouseResponse(result.House)}
}

func toListItemResponse(item housesvc.ListItem) listItemResponse {
	house := item.HouseSummary
	return listItemResponse{
		ListingID:     house.ListingID,
		AssetMode:     house.AssetMode,
		ProviderID:    house.ProviderID,
		ProviderPhone: house.ProviderPhone,
		Title:         house.Title,
		City:          house.City,
		District:      house.District,
		BizArea:       house.BizArea,
		CommunityName: house.CommunityName,
		BuildingName:  house.BuildingName,
		RoomNo:        house.RoomNo,
		Price:         house.Price,
		PriceText:     house.PriceText,
		LayoutText:    house.LayoutText,
		AreaSize:      house.AreaSize,
		RoomStatus:    house.RoomStatus,
		ListingStatus: house.ListingStatus,
		AuditStatus:   house.AuditStatus,
		IsOnline:      house.IsOnline,
		UpdatedAt:     house.UpdatedAt,
	}
}

func toHouseResponse(house housesvc.HouseSummary) houseResponse {
	return houseResponse{
		ListingID:         house.ListingID,
		SourceType:        house.SourceType,
		SourceID:          house.SourceID,
		AssetMode:         house.AssetMode,
		ProviderID:        house.ProviderID,
		ProviderPhone:     house.ProviderPhone,
		LandlordName:      house.LandlordName,
		RootType:          house.RootType,
		RootID:            house.RootID,
		ProjectID:         house.ProjectID,
		ProjectName:       house.ProjectName,
		BuildingID:        house.BuildingID,
		BuildingName:      house.BuildingName,
		RoomTypeID:        house.RoomTypeID,
		RoomTypeName:      house.RoomTypeName,
		DecentralizedID:   house.DecentralizedID,
		CommunityName:     house.CommunityName,
		RentMode:          house.RentMode,
		City:              house.City,
		District:          house.District,
		BizArea:           house.BizArea,
		AddressText:       house.AddressText,
		RoomNo:            house.RoomNo,
		Title:             house.Title,
		Price:             house.Price,
		PriceText:         house.PriceText,
		LayoutText:        house.LayoutText,
		AreaSize:          house.AreaSize,
		RoomStatus:        house.RoomStatus,
		ListingStatus:     house.ListingStatus,
		AuditStatus:       house.AuditStatus,
		IsOnline:          house.IsOnline,
		LatestAuditTaskID: house.LatestAuditTaskID,
		LatestSubmittedAt: house.LatestSubmittedAt,
		LatestReviewedAt:  house.LatestReviewedAt,
		ReviewerStaffID:   house.ReviewerStaffID,
		CreatedAt:         house.CreatedAt,
		UpdatedAt:         house.UpdatedAt,
	}
}
