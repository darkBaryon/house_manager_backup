package listingprojection

import (
	"fmt"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	"strings"
)

func mapCentralizedMiniappListing(
	listing *hpdmodel.HpdListing,
	room *hmdmodel.HmdRoomCentralized,
	project *hmdmodel.HmdCentralized,
	building *hmdmodel.HmdBuilding,
	roomType *hmdmodel.HmdRoomTypeCentralized,
) *hpdmodel.HpdMiniappListing {
	layoutText := firstNonBlank(room.LayoutText, roomTypeLayoutText(roomType))
	roomCount, hallCount, bathroomCount, kitchenCount := centralizedProjectionShape(room, roomType)
	areaSize := firstPositiveInt(room.AreaSize, roomTypeAreaSize(roomType))
	orientation := firstOrientation(room.Orientation, roomTypeOrientation(roomType))
	paymentCycle := firstPaymentCycle(room.PaymentCycle, roomTypePaymentCycle(roomType))
	images := firstImages(room.Images, roomTypeImages(roomType))

	return &hpdmodel.HpdMiniappListing{
		ListingID:               listing.ID,
		SourceType:              listing.SourceType,
		SourceID:                listing.SourceID,
		AssetMode:               listing.AssetMode,
		RentMode:                room.RentMode,
		City:                    project.City,
		District:                project.District,
		CommunityName:           project.ProjectName,
		BuildingOrCommunityName: firstNonBlank(building.BuildingName, project.ProjectName),
		AddressText:             project.AddressText,
		Geo:                     project.Geo,
		Title:                   joinNonBlank(" ", project.ProjectName, building.BuildingName, room.RoomNo),
		Subtitle:                roomSubtitle(layoutText, areaSize, orientation, room.FloorNo),
		Price:                   room.Rent,
		PriceText:               priceText(room.Rent),
		LayoutText:              layoutText,
		RoomCount:               roomCount,
		HallCount:               hallCount,
		BathroomCount:           bathroomCount,
		KitchenCount:            kitchenCount,
		AreaSize:                areaSize,
		Orientation:             orientation,
		FloorText:               floorText(room.FloorNo),
		PaymentCycle:            paymentCycle,
		FeatureFlags:            projectionFeatureFlags(room.RoomFacilities, room.ListingFacilities, building.ListingFacilities),
		ListingFacilities:       mergeListingFacilities(room.ListingFacilities, building.ListingFacilities),
		StartRentRule:           room.StartRentRule,
		Images:                  images,
		IsOnline:                hpdmodel.HpdMiniappOnlineStatus(listing.ListingStatus, room.RoomStatus),
	}
}

func mapDecentralizedMiniappListing(
	listing *hpdmodel.HpdListing,
	room *hmdmodel.HmdRoomDecentralized,
	community *hmdmodel.HmdDecentralized,
) *hpdmodel.HpdMiniappListing {
	return &hpdmodel.HpdMiniappListing{
		ListingID:               listing.ID,
		SourceType:              listing.SourceType,
		SourceID:                listing.SourceID,
		AssetMode:               listing.AssetMode,
		RentMode:                room.RentMode,
		City:                    community.City,
		District:                community.District,
		BizArea:                 community.BizArea,
		CommunityName:           community.CommunityName,
		BuildingOrCommunityName: community.CommunityName,
		SubwayStation:           community.SubwayStation,
		AddressText:             community.AddressText,
		Geo:                     community.Geo,
		Title:                   joinNonBlank(" ", community.CommunityName, room.RoomNo),
		Subtitle:                roomSubtitle(room.LayoutText, room.AreaSize, room.Orientation, room.FloorNo),
		Price:                   room.Rent,
		PriceText:               priceText(room.Rent),
		LayoutText:              room.LayoutText,
		RoomCount:               room.RoomCount,
		HallCount:               room.HallCount,
		BathroomCount:           room.BathroomCount,
		KitchenCount:            room.KitchenCount,
		AreaSize:                room.AreaSize,
		Orientation:             room.Orientation,
		FloorText:               floorText(room.FloorNo),
		PaymentCycle:            room.PaymentCycle,
		FeatureFlags:            projectionFeatureFlags(room.RoomFacilities, room.ListingFacilities),
		ListingFacilities:       cloneListingFacilities(room.ListingFacilities),
		StartRentRule:           room.StartRentRule,
		Images:                  cloneImages(room.Images),
		IsOnline:                hpdmodel.HpdMiniappOnlineStatus(listing.ListingStatus, room.RoomStatus),
	}
}

func centralizedProjectionShape(room *hmdmodel.HmdRoomCentralized, roomType *hmdmodel.HmdRoomTypeCentralized) (int, int, int, int) {
	if roomType == nil {
		return layoutCountFromRoom(room, func(r *hmdmodel.HmdRoomCentralized) int { return r.RoomCount }),
			layoutCountFromRoom(room, func(r *hmdmodel.HmdRoomCentralized) int { return r.HallCount }),
			layoutCountFromRoom(room, func(r *hmdmodel.HmdRoomCentralized) int { return r.BathroomCount }),
			layoutCountFromRoom(room, func(r *hmdmodel.HmdRoomCentralized) int { return r.KitchenCount })
	}
	return layoutCountWithFallback(room, roomType.RoomCount, func(r *hmdmodel.HmdRoomCentralized) int { return r.RoomCount }),
		layoutCountWithFallback(room, roomType.HallCount, func(r *hmdmodel.HmdRoomCentralized) int { return r.HallCount }),
		layoutCountWithFallback(room, roomType.BathroomCount, func(r *hmdmodel.HmdRoomCentralized) int { return r.BathroomCount }),
		layoutCountWithFallback(room, roomType.KitchenCount, func(r *hmdmodel.HmdRoomCentralized) int { return r.KitchenCount })
}

func layoutCountFromRoom(room *hmdmodel.HmdRoomCentralized, pick func(*hmdmodel.HmdRoomCentralized) int) int {
	if room == nil {
		return hmdmodel.UnknownLayoutCount
	}
	return pick(room)
}

func layoutCountWithFallback(room *hmdmodel.HmdRoomCentralized, fallback int, pick func(*hmdmodel.HmdRoomCentralized) int) int {
	value := layoutCountFromRoom(room, pick)
	if value != hmdmodel.UnknownLayoutCount {
		return value
	}
	return fallback
}

func roomTypeLayoutText(roomType *hmdmodel.HmdRoomTypeCentralized) string {
	if roomType == nil {
		return ""
	}
	if roomType.RoomCount <= 0 {
		return roomType.RoomTypeName
	}
	text := fmt.Sprintf("%d室", roomType.RoomCount)
	if roomType.HallCount > 0 {
		text += fmt.Sprintf("%d厅", roomType.HallCount)
	}
	if roomType.BathroomCount > 0 {
		text += fmt.Sprintf("%d卫", roomType.BathroomCount)
	}
	return text
}

func roomTypeAreaSize(roomType *hmdmodel.HmdRoomTypeCentralized) int {
	if roomType == nil {
		return 0
	}
	return roomType.AreaSize
}

func roomTypeOrientation(roomType *hmdmodel.HmdRoomTypeCentralized) hmdmodel.Orientation {
	if roomType == nil {
		return ""
	}
	return roomType.Orientation
}

func roomTypePaymentCycle(roomType *hmdmodel.HmdRoomTypeCentralized) hmdmodel.PaymentCycle {
	if roomType == nil {
		return ""
	}
	return roomType.PaymentCycle
}

func roomTypeDecorationLevel(roomType *hmdmodel.HmdRoomTypeCentralized) hmdmodel.DecorationLevel {
	if roomType == nil {
		return ""
	}
	return roomType.DecorationLevel
}

func roomTypeAgencyFeeMode(roomType *hmdmodel.HmdRoomTypeCentralized) hmdmodel.AgencyFeeMode {
	if roomType == nil {
		return ""
	}
	return roomType.AgencyFeeMode
}

func roomTypeAgencyFeeValue(roomType *hmdmodel.HmdRoomTypeCentralized) int {
	if roomType == nil {
		return 0
	}
	return roomType.AgencyFeeValue
}

func roomTypeRent(roomType *hmdmodel.HmdRoomTypeCentralized) int {
	if roomType == nil {
		return 0
	}
	return roomType.Rent
}

func roomTypeDeposit(roomType *hmdmodel.HmdRoomTypeCentralized) int {
	if roomType == nil {
		return 0
	}
	return roomType.Deposit
}

func roomTypeServiceFee(roomType *hmdmodel.HmdRoomTypeCentralized) int {
	if roomType == nil {
		return 0
	}
	return roomType.ServiceFee
}

func roomTypeName(roomType *hmdmodel.HmdRoomTypeCentralized) string {
	if roomType == nil {
		return ""
	}
	return roomType.RoomTypeName
}

func roomTypeImages(roomType *hmdmodel.HmdRoomTypeCentralized) []hmdmodel.TaggedImage {
	if roomType == nil {
		return nil
	}
	return roomType.Images
}

func roomSubtitle(layoutText string, areaSize int, orientation hmdmodel.Orientation, floorNo int) string {
	parts := make([]string, 0, 4)
	if strings.TrimSpace(layoutText) != "" {
		parts = append(parts, strings.TrimSpace(layoutText))
	}
	if areaSize > 0 {
		parts = append(parts, fmt.Sprintf("%d㎡", areaSize))
	}
	if orientation != "" {
		parts = append(parts, string(orientation))
	}
	if floorNo > 0 {
		parts = append(parts, floorText(floorNo))
	}
	return strings.Join(parts, " | ")
}

func floorText(floorNo int) string {
	if floorNo <= 0 {
		return ""
	}
	return fmt.Sprintf("%d层", floorNo)
}

func priceText(price int) string {
	if price <= 0 {
		return ""
	}
	return fmt.Sprintf("%d元/月", price)
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstPositiveInt(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func firstOrientation(values ...hmdmodel.Orientation) hmdmodel.Orientation {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstPaymentCycle(values ...hmdmodel.PaymentCycle) hmdmodel.PaymentCycle {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstDecorationLevel(values ...hmdmodel.DecorationLevel) hmdmodel.DecorationLevel {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstAgencyFeeMode(values ...hmdmodel.AgencyFeeMode) hmdmodel.AgencyFeeMode {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstImages(values ...[]hmdmodel.TaggedImage) []hmdmodel.TaggedImage {
	for _, value := range values {
		if len(value) > 0 {
			return cloneImages(value)
		}
	}
	return nil
}

func cloneImages(items []hmdmodel.TaggedImage) []hmdmodel.TaggedImage {
	if len(items) == 0 {
		return nil
	}
	out := make([]hmdmodel.TaggedImage, len(items))
	copy(out, items)
	return out
}

func cloneListingFacilities(items []hmdmodel.ListingFacility) []hmdmodel.ListingFacility {
	if len(items) == 0 {
		return nil
	}
	out := make([]hmdmodel.ListingFacility, len(items))
	copy(out, items)
	return out
}

func cloneRoomFacilities(items []hmdmodel.RoomFacility) []hmdmodel.RoomFacility {
	if len(items) == 0 {
		return nil
	}
	out := make([]hmdmodel.RoomFacility, len(items))
	copy(out, items)
	return out
}

func mergeListingFacilities(items ...[]hmdmodel.ListingFacility) []hmdmodel.ListingFacility {
	seen := map[hmdmodel.ListingFacility]struct{}{}
	out := []hmdmodel.ListingFacility{}
	for _, group := range items {
		for _, item := range group {
			if item == "" {
				continue
			}
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			out = append(out, item)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func projectionFeatureFlags(roomFacilities []hmdmodel.RoomFacility, listingFacilities ...[]hmdmodel.ListingFacility) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, item := range roomFacilities {
		value := string(item)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	for _, group := range listingFacilities {
		for _, item := range group {
			value := string(item)
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			out = append(out, value)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func joinNonBlank(sep string, values ...string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			parts = append(parts, strings.TrimSpace(value))
		}
	}
	return strings.Join(parts, sep)
}
