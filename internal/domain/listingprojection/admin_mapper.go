package listingprojection

import (
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
)

func mapCentralizedAdminListing(
	listing *hpdmodel.HpdListing,
	room *hmdmodel.HmdRoomCentralized,
	project *hmdmodel.HmdCentralized,
	building *hmdmodel.HmdBuilding,
	roomType *hmdmodel.HmdRoomTypeCentralized,
	owner *hpdmodel.HpdRootScopeRelation,
) *hpdmodel.HpdAdminListing {
	publisher := mapCentralizedPublisherListing(listing, room, project, building, roomType, owner)
	return adminListingFromPublisher(publisher)
}

func mapDecentralizedAdminListing(
	listing *hpdmodel.HpdListing,
	room *hmdmodel.HmdRoomDecentralized,
	community *hmdmodel.HmdDecentralized,
	owner *hpdmodel.HpdRootScopeRelation,
) *hpdmodel.HpdAdminListing {
	publisher := mapDecentralizedPublisherListing(listing, room, community, owner)
	return adminListingFromPublisher(publisher)
}

func adminListingFromPublisher(publisher *hpdmodel.HpdPublisherListing) *hpdmodel.HpdAdminListing {
	if publisher == nil {
		return nil
	}
	return &hpdmodel.HpdAdminListing{
		ListingID:            publisher.ListingID,
		SourceType:           publisher.SourceType,
		SourceID:             publisher.SourceID,
		AssetMode:            publisher.AssetMode,
		OwnerLandlordID:      publisher.OwnerLandlordID,
		OwnerPhoneSnapshot:   publisher.OwnerPhoneSnapshot,
		LandlordNameSnapshot: publisher.LandlordNameSnapshot,
		RootType:             publisher.RootType,
		RootID:               publisher.RootID,
		ProjectID:            publisher.ProjectID,
		ProjectName:          publisher.ProjectName,
		BuildingID:           publisher.BuildingID,
		BuildingName:         publisher.BuildingName,
		RoomTypeID:           publisher.RoomTypeID,
		RoomTypeName:         publisher.RoomTypeName,
		DecentralizedID:      publisher.DecentralizedID,
		CommunityName:        publisher.CommunityName,
		RentMode:             publisher.RentMode,
		City:                 publisher.City,
		District:             publisher.District,
		BizArea:              publisher.BizArea,
		AddressText:          publisher.AddressText,
		RoomNo:               publisher.RoomNo,
		Title:                publisher.Title,
		Price:                publisher.Price,
		PriceText:            publisher.PriceText,
		LayoutText:           publisher.LayoutText,
		RoomCount:            publisher.RoomCount,
		HallCount:            publisher.HallCount,
		BathroomCount:        publisher.BathroomCount,
		KitchenCount:         publisher.KitchenCount,
		AreaSize:             publisher.AreaSize,
		RoomStatus:           publisher.RoomStatus,
		ListingStatus:        publisher.ListingStatus,
		AuditStatus:          hpdmodel.HpdAuditStatusUnspecified,
		IsOnline:             publisher.IsOnline,
	}
}
