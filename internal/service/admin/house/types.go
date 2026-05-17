package house

type HouseSummary struct {
	ListingID         string
	SourceType        string
	SourceID          string
	AssetMode         string
	ProviderID        string
	ProviderPhone     string
	LandlordName      string
	RootType          string
	RootID            string
	ProjectID         string
	ProjectName       string
	BuildingID        string
	BuildingName      string
	RoomTypeID        string
	RoomTypeName      string
	DecentralizedID   string
	CommunityName     string
	RentMode          string
	City              string
	District          string
	BizArea           string
	AddressText       string
	RoomNo            string
	Title             string
	Price             int
	PriceText         string
	LayoutText        string
	AreaSize          int
	RoomStatus        int
	ListingStatus     int
	AuditStatus       int
	IsOnline          int
	LatestAuditTaskID string
	LatestSubmittedAt int64
	LatestReviewedAt  int64
	ReviewerStaffID   string
	CreatedAt         int64
	UpdatedAt         int64
}

type RootSummary struct {
	RootID        string
	RootType      string
	RootName      string
	AssetMode     string
	ProviderID    string
	ProviderPhone string
	ProviderName  string
	ProjectID     string
	ProjectName   string
	CommunityID   string
	CommunityName string
	City          string
	District      string
	BizArea       string
	BuildingCount int
	RoomCount     int64
	UpdatedAt     int64
}

type RootListInput struct {
	ProviderID    string
	AssetMode     string
	City          string
	District      string
	RoomStatus    *int
	ListingStatus *int
	AuditStatus   *int
	Page          int
	PageSize      int
}

type RootListItem struct {
	RootSummary
}

type RootListResult struct {
	List     []RootListItem
	Page     int
	PageSize int
	Total    int64
}

type BuildingSummary struct {
	RootID       string
	BuildingID   string
	ProjectID    string
	ProjectName  string
	BuildingName string
	City         string
	District     string
	BizArea      string
	RoomCount    int64
	UpdatedAt    int64
}

type BuildingListInput struct {
	RootID        string
	RoomStatus    *int
	ListingStatus *int
	AuditStatus   *int
	Page          int
	PageSize      int
}

type BuildingListItem struct {
	BuildingSummary
}

type BuildingListResult struct {
	List     []BuildingListItem
	Page     int
	PageSize int
	Total    int64
}

type ListInput struct {
	RootID        string
	BuildingID    string
	ProviderID    string
	AssetMode     string
	City          string
	District      string
	RoomStatus    *int
	ListingStatus *int
	AuditStatus   *int
	Page          int
	PageSize      int
}

type ListItem struct {
	HouseSummary
}

type ListResult struct {
	List     []ListItem
	Page     int
	PageSize int
	Total    int64
}

type DetailInput struct {
	ListingID string
}

type DetailResult struct {
	House HouseSummary
}
