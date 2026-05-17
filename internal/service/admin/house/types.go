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

type ListInput struct {
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
