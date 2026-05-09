package hpd

import (
	"fmt"
	"strings"

	"house-manager/internal/model"
)

func mapCentralizedMiniappListing(
	listing *model.HpdListing,
	room *model.HmdRoomCentralized,
	project *model.HmdCentralized,
	building *model.HmdBuilding,
	roomType *model.HmdRoomTypeCentralized,
) *model.HpdMiniappListing {
	layoutText := firstNonBlank(room.LayoutText, roomTypeLayoutText(roomType))
	areaSize := firstPositiveInt(room.AreaSize, roomTypeAreaSize(roomType))
	orientation := firstOrientation(room.Orientation, roomTypeOrientation(roomType))
	paymentCycle := firstPaymentCycle(room.PaymentCycle, roomTypePaymentCycle(roomType))
	images := firstImages(room.Images, roomTypeImages(roomType))

	return &model.HpdMiniappListing{
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
		AreaSize:                areaSize,
		Orientation:             orientation,
		FloorText:               floorText(room.FloorNo),
		PaymentCycle:            paymentCycle,
		FeatureFlags:            projectionFeatureFlags(room.RoomFacilities, room.ListingFacilities, building.ListingFacilities),
		ListingFacilities:       mergeListingFacilities(room.ListingFacilities, building.ListingFacilities),
		StartRentRule:           room.StartRentRule,
		Images:                  images,
		IsOnline:                model.HpdMiniappOnlineStatus(listing.ListingStatus, room.RoomStatus),
	}
}

func mapDecentralizedMiniappListing(
	listing *model.HpdListing,
	room *model.HmdRoomDecentralized,
	community *model.HmdDecentralized,
) *model.HpdMiniappListing {
	return &model.HpdMiniappListing{
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
		AreaSize:                room.AreaSize,
		Orientation:             room.Orientation,
		FloorText:               floorText(room.FloorNo),
		PaymentCycle:            room.PaymentCycle,
		FeatureFlags:            projectionFeatureFlags(room.RoomFacilities, room.ListingFacilities),
		ListingFacilities:       cloneListingFacilities(room.ListingFacilities),
		StartRentRule:           room.StartRentRule,
		Images:                  cloneImages(room.Images),
		IsOnline:                model.HpdMiniappOnlineStatus(listing.ListingStatus, room.RoomStatus),
	}
}

func roomTypeLayoutText(roomType *model.HmdRoomTypeCentralized) string {
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

func roomTypeAreaSize(roomType *model.HmdRoomTypeCentralized) int {
	if roomType == nil {
		return 0
	}
	return roomType.AreaSize
}

func roomTypeOrientation(roomType *model.HmdRoomTypeCentralized) model.Orientation {
	if roomType == nil {
		return ""
	}
	return roomType.Orientation
}

func roomTypePaymentCycle(roomType *model.HmdRoomTypeCentralized) model.PaymentCycle {
	if roomType == nil {
		return ""
	}
	return roomType.PaymentCycle
}

func roomTypeImages(roomType *model.HmdRoomTypeCentralized) []model.TaggedImage {
	if roomType == nil {
		return nil
	}
	return roomType.Images
}

func roomSubtitle(layoutText string, areaSize int, orientation model.Orientation, floorNo int) string {
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

func firstOrientation(values ...model.Orientation) model.Orientation {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstPaymentCycle(values ...model.PaymentCycle) model.PaymentCycle {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstImages(values ...[]model.TaggedImage) []model.TaggedImage {
	for _, value := range values {
		if len(value) > 0 {
			return cloneImages(value)
		}
	}
	return nil
}

func cloneImages(items []model.TaggedImage) []model.TaggedImage {
	if len(items) == 0 {
		return nil
	}
	out := make([]model.TaggedImage, len(items))
	copy(out, items)
	return out
}

func cloneListingFacilities(items []model.ListingFacility) []model.ListingFacility {
	if len(items) == 0 {
		return nil
	}
	out := make([]model.ListingFacility, len(items))
	copy(out, items)
	return out
}

func mergeListingFacilities(items ...[]model.ListingFacility) []model.ListingFacility {
	seen := map[model.ListingFacility]struct{}{}
	out := []model.ListingFacility{}
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

func projectionFeatureFlags(roomFacilities []model.RoomFacility, listingFacilities ...[]model.ListingFacility) []string {
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
