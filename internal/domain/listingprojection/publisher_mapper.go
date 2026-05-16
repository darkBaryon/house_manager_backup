package listingprojection

import (
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func mapCentralizedPublisherListing(
	listing *hpdmodel.HpdListing,
	room *hmdmodel.HmdRoomCentralized,
	project *hmdmodel.HmdCentralized,
	building *hmdmodel.HmdBuilding,
	roomType *hmdmodel.HmdRoomTypeCentralized,
	owner *hpdmodel.HpdRootScopeRelation,
) *hpdmodel.HpdPublisherListing {
	layoutText := firstNonBlank(room.LayoutText, roomTypeLayoutText(roomType))
	areaSize := firstPositiveInt(room.AreaSize, roomTypeAreaSize(roomType))
	orientation := firstOrientation(room.Orientation, roomTypeOrientation(roomType))
	paymentCycle := firstPaymentCycle(room.PaymentCycle, roomTypePaymentCycle(roomType))
	decorationLevel := firstDecorationLevel(room.DecorationLevel, roomTypeDecorationLevel(roomType))
	images := firstImages(room.Images, roomTypeImages(roomType))
	price := firstPositiveInt(room.Rent, roomTypeRent(roomType))
	deposit := firstPositiveInt(room.Deposit, roomTypeDeposit(roomType))
	serviceFee := firstPositiveInt(room.ServiceFee, roomTypeServiceFee(roomType))
	agencyFeeValue := firstPositiveInt(room.AgencyFeeValue, roomTypeAgencyFeeValue(roomType))
	agencyFeeMode := firstAgencyFeeMode(room.AgencyFeeMode, roomTypeAgencyFeeMode(roomType))

	return &hpdmodel.HpdPublisherListing{
		ListingID:          listing.ID,
		SourceType:         listing.SourceType,
		SourceID:           listing.SourceID,
		AssetMode:          listing.AssetMode,
		OwnerLandlordID:    ownerLandlordID(owner),
		OwnerPhoneSnapshot: ownerPhone(owner),
		RootType:           hpdmodel.HpdRootScopeTypeCentralizedProject,
		RootID:             project.ID,
		ProjectID:          project.ID,
		ProjectName:        project.ProjectName,
		BuildingID:         building.ID,
		BuildingName:       building.BuildingName,
		RoomTypeID:         room.RoomTypeID,
		RoomTypeName:       roomTypeName(roomType),
		RentMode:           room.RentMode,
		City:               project.City,
		District:           project.District,
		AddressText:        project.AddressText,
		Geo:                project.Geo,
		RoomNo:             room.RoomNo,
		FloorNo:            room.FloorNo,
		Title:              joinNonBlank(" ", project.ProjectName, building.BuildingName, room.RoomNo),
		Subtitle:           roomSubtitle(layoutText, areaSize, orientation, room.FloorNo),
		Price:              price,
		PriceText:          priceText(price),
		LayoutText:         layoutText,
		AreaSize:           areaSize,
		Orientation:        orientation,
		DecorationLevel:    decorationLevel,
		PaymentCycle:       paymentCycle,
		Deposit:            deposit,
		ServiceFee:         serviceFee,
		AgencyFeeMode:      agencyFeeMode,
		AgencyFeeValue:     agencyFeeValue,
		RoomStatus:         room.RoomStatus,
		ListingStatus:      listing.ListingStatus,
		ViewingTimeRule:    room.ViewingTimeRule,
		StartRentRule:      room.StartRentRule,
		FeatureFlags:       projectionFeatureFlags(room.RoomFacilities, room.ListingFacilities, building.ListingFacilities),
		ListingFacilities:  mergeListingFacilities(room.ListingFacilities, building.ListingFacilities),
		RoomFacilities:     cloneRoomFacilities(room.RoomFacilities),
		Images:             images,
		IsOnline:           hpdmodel.HpdMiniappOnlineStatus(listing.ListingStatus, room.RoomStatus),
	}
}

func mapDecentralizedPublisherListing(
	listing *hpdmodel.HpdListing,
	room *hmdmodel.HmdRoomDecentralized,
	community *hmdmodel.HmdDecentralized,
	owner *hpdmodel.HpdRootScopeRelation,
) *hpdmodel.HpdPublisherListing {
	return &hpdmodel.HpdPublisherListing{
		ListingID:          listing.ID,
		SourceType:         listing.SourceType,
		SourceID:           listing.SourceID,
		AssetMode:          listing.AssetMode,
		OwnerLandlordID:    ownerLandlordID(owner),
		OwnerPhoneSnapshot: ownerPhone(owner),
		RootType:           hpdmodel.HpdRootScopeTypeDecentralizedCommunity,
		RootID:             community.ID,
		DecentralizedID:    community.ID,
		CommunityName:      community.CommunityName,
		RentMode:           room.RentMode,
		City:               community.City,
		District:           community.District,
		BizArea:            community.BizArea,
		SubwayStation:      community.SubwayStation,
		AddressText:        community.AddressText,
		Geo:                community.Geo,
		RoomNo:             room.RoomNo,
		FloorNo:            room.FloorNo,
		Title:              joinNonBlank(" ", community.CommunityName, room.RoomNo),
		Subtitle:           roomSubtitle(room.LayoutText, room.AreaSize, room.Orientation, room.FloorNo),
		Price:              room.Rent,
		PriceText:          priceText(room.Rent),
		LayoutText:         room.LayoutText,
		AreaSize:           room.AreaSize,
		Orientation:        room.Orientation,
		DecorationLevel:    room.DecorationLevel,
		PaymentCycle:       room.PaymentCycle,
		Deposit:            room.Deposit,
		ServiceFee:         room.ServiceFee,
		AgencyFeeMode:      room.AgencyFeeMode,
		AgencyFeeValue:     room.AgencyFeeValue,
		RoomStatus:         room.RoomStatus,
		ListingStatus:      listing.ListingStatus,
		ViewingTimeRule:    room.ViewingTimeRule,
		StartRentRule:      room.StartRentRule,
		FeatureFlags:       projectionFeatureFlags(room.RoomFacilities, room.ListingFacilities),
		ListingFacilities:  cloneListingFacilities(room.ListingFacilities),
		RoomFacilities:     cloneRoomFacilities(room.RoomFacilities),
		Images:             cloneImages(room.Images),
		IsOnline:           hpdmodel.HpdMiniappOnlineStatus(listing.ListingStatus, room.RoomStatus),
	}
}

func ownerLandlordID(owner *hpdmodel.HpdRootScopeRelation) bson.ObjectID {
	if owner == nil {
		return bson.NilObjectID
	}
	return owner.OwnerLandlordID
}

func ownerPhone(owner *hpdmodel.HpdRootScopeRelation) string {
	if owner == nil {
		return ""
	}
	return owner.OwnerPhone
}
