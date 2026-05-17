package house

import housesvc "house-manager/internal/service/admin/house"

type rootListRequest struct {
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

type buildingListRequest struct {
	RootID        string `json:"root_id"`
	RoomStatus    *int   `json:"room_status"`
	ListingStatus *int   `json:"listing_status"`
	AuditStatus   *int   `json:"audit_status"`
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
}

type roomListRequest struct {
	RootID        string `json:"root_id"`
	BuildingID    string `json:"building_id"`
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

type rootListItemResponse struct {
	RootID        string `json:"root_id"`
	RootType      string `json:"root_type"`
	RootName      string `json:"root_name"`
	AssetMode     string `json:"asset_mode"`
	ProviderID    string `json:"provider_id"`
	ProviderPhone string `json:"provider_phone"`
	ProviderName  string `json:"provider_name"`
	ProjectID     string `json:"project_id"`
	ProjectName   string `json:"project_name"`
	CommunityID   string `json:"community_id"`
	CommunityName string `json:"community_name"`
	City          string `json:"city"`
	District      string `json:"district"`
	BizArea       string `json:"biz_area"`
	BuildingCount int    `json:"building_count"`
	RoomCount     int64  `json:"room_count"`
	UpdatedAt     int64  `json:"updated_at"`
}

type rootListResponse struct {
	List     []rootListItemResponse `json:"list"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Total    int64                  `json:"total"`
}

type buildingListItemResponse struct {
	RootID       string `json:"root_id"`
	BuildingID   string `json:"building_id"`
	ProjectID    string `json:"project_id"`
	ProjectName  string `json:"project_name"`
	BuildingName string `json:"building_name"`
	City         string `json:"city"`
	District     string `json:"district"`
	BizArea      string `json:"biz_area"`
	RoomCount    int64  `json:"room_count"`
	UpdatedAt    int64  `json:"updated_at"`
}

type buildingListResponse struct {
	List     []buildingListItemResponse `json:"list"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"page_size"`
	Total    int64                      `json:"total"`
}

type roomListItemResponse struct {
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

type roomListResponse struct {
	List     []roomListItemResponse `json:"list"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Total    int64                  `json:"total"`
}

type detailResponse struct {
	Room roomResponse `json:"room"`
}

type roomResponse struct {
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

func toRootListResponse(result *housesvc.RootListResult) rootListResponse {
	items := make([]rootListItemResponse, 0, len(result.List))
	for _, item := range result.List {
		root := item.RootSummary
		items = append(items, rootListItemResponse{
			RootID:        root.RootID,
			RootType:      root.RootType,
			RootName:      root.RootName,
			AssetMode:     root.AssetMode,
			ProviderID:    root.ProviderID,
			ProviderPhone: root.ProviderPhone,
			ProviderName:  root.ProviderName,
			ProjectID:     root.ProjectID,
			ProjectName:   root.ProjectName,
			CommunityID:   root.CommunityID,
			CommunityName: root.CommunityName,
			City:          root.City,
			District:      root.District,
			BizArea:       root.BizArea,
			BuildingCount: root.BuildingCount,
			RoomCount:     root.RoomCount,
			UpdatedAt:     root.UpdatedAt,
		})
	}
	if items == nil {
		items = []rootListItemResponse{}
	}
	return rootListResponse{
		List:     items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}
}

func toBuildingListResponse(result *housesvc.BuildingListResult) buildingListResponse {
	items := make([]buildingListItemResponse, 0, len(result.List))
	for _, item := range result.List {
		building := item.BuildingSummary
		items = append(items, buildingListItemResponse{
			RootID:       building.RootID,
			BuildingID:   building.BuildingID,
			ProjectID:    building.ProjectID,
			ProjectName:  building.ProjectName,
			BuildingName: building.BuildingName,
			City:         building.City,
			District:     building.District,
			BizArea:      building.BizArea,
			RoomCount:    building.RoomCount,
			UpdatedAt:    building.UpdatedAt,
		})
	}
	if items == nil {
		items = []buildingListItemResponse{}
	}
	return buildingListResponse{
		List:     items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}
}

func toRoomListResponse(result *housesvc.ListResult) roomListResponse {
	items := make([]roomListItemResponse, 0, len(result.List))
	for _, item := range result.List {
		room := item.HouseSummary
		items = append(items, roomListItemResponse{
			ListingID:     room.ListingID,
			AssetMode:     room.AssetMode,
			ProviderID:    room.ProviderID,
			ProviderPhone: room.ProviderPhone,
			Title:         room.Title,
			City:          room.City,
			District:      room.District,
			BizArea:       room.BizArea,
			CommunityName: room.CommunityName,
			BuildingName:  room.BuildingName,
			RoomNo:        room.RoomNo,
			Price:         room.Price,
			PriceText:     room.PriceText,
			LayoutText:    room.LayoutText,
			AreaSize:      room.AreaSize,
			RoomStatus:    room.RoomStatus,
			ListingStatus: room.ListingStatus,
			AuditStatus:   room.AuditStatus,
			IsOnline:      room.IsOnline,
			UpdatedAt:     room.UpdatedAt,
		})
	}
	if items == nil {
		items = []roomListItemResponse{}
	}
	return roomListResponse{
		List:     items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}
}

func toDetailResponse(result *housesvc.DetailResult) detailResponse {
	return detailResponse{Room: toRoomResponse(result.House)}
}

func toRoomResponse(room housesvc.HouseSummary) roomResponse {
	return roomResponse{
		ListingID:         room.ListingID,
		SourceType:        room.SourceType,
		SourceID:          room.SourceID,
		AssetMode:         room.AssetMode,
		ProviderID:        room.ProviderID,
		ProviderPhone:     room.ProviderPhone,
		LandlordName:      room.LandlordName,
		RootType:          room.RootType,
		RootID:            room.RootID,
		ProjectID:         room.ProjectID,
		ProjectName:       room.ProjectName,
		BuildingID:        room.BuildingID,
		BuildingName:      room.BuildingName,
		RoomTypeID:        room.RoomTypeID,
		RoomTypeName:      room.RoomTypeName,
		DecentralizedID:   room.DecentralizedID,
		CommunityName:     room.CommunityName,
		RentMode:          room.RentMode,
		City:              room.City,
		District:          room.District,
		BizArea:           room.BizArea,
		AddressText:       room.AddressText,
		RoomNo:            room.RoomNo,
		Title:             room.Title,
		Price:             room.Price,
		PriceText:         room.PriceText,
		LayoutText:        room.LayoutText,
		AreaSize:          room.AreaSize,
		RoomStatus:        room.RoomStatus,
		ListingStatus:     room.ListingStatus,
		AuditStatus:       room.AuditStatus,
		IsOnline:          room.IsOnline,
		LatestAuditTaskID: room.LatestAuditTaskID,
		LatestSubmittedAt: room.LatestSubmittedAt,
		LatestReviewedAt:  room.LatestReviewedAt,
		ReviewerStaffID:   room.ReviewerStaffID,
		CreatedAt:         room.CreatedAt,
		UpdatedAt:         room.UpdatedAt,
	}
}
