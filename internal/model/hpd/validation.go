package hpd

import (
	"fmt"

	commonmodel "house-manager/internal/model/common"
	hmdmodel "house-manager/internal/model/hmd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (m *HpdListing) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("hpd listing is nil")
	}
	if m.SourceID.IsZero() {
		return fmt.Errorf("sourceID is required")
	}
	if !m.SourceType.Valid() {
		return fmt.Errorf("sourceType is invalid")
	}
	if !m.AssetMode.Valid() {
		return fmt.Errorf("assetMode is invalid")
	}
	if !HpdSourceAssetModeMatch(m.SourceType, m.AssetMode) {
		return fmt.Errorf("sourceType and assetMode mismatch")
	}
	if m.ListingStatus == HpdListingStatusUnspecified || !m.ListingStatus.Valid() {
		return fmt.Errorf("listingStatus is invalid")
	}
	if err := validateNonNegativeInt64("publishedAt", m.PublishedAt); err != nil {
		return err
	}
	return validateNonNegativeInt64("offlineAt", m.OfflineAt)
}

func (m *HpdMiniappListing) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("hpd miniapp listing is nil")
	}
	if m.ListingID.IsZero() || m.SourceID.IsZero() {
		return fmt.Errorf("listingID and sourceID are required")
	}
	if !m.SourceType.Valid() {
		return fmt.Errorf("sourceType is invalid")
	}
	if !m.AssetMode.Valid() {
		return fmt.Errorf("assetMode is invalid")
	}
	if !HpdSourceAssetModeMatch(m.SourceType, m.AssetMode) {
		return fmt.Errorf("sourceType and assetMode mismatch")
	}
	if !m.RentMode.Valid() {
		return fmt.Errorf("rentMode is invalid")
	}
	if commonmodel.IsBlank(m.City) || commonmodel.IsBlank(m.Title) {
		return fmt.Errorf("city and title are required")
	}
	return validateHpdMiniappListingFields(
		m.Price,
		m.SubwayDistanceM,
		m.AreaSize,
		m.WeightScore,
		m.Orientation,
		m.PaymentCycle,
		m.StartRentRule,
		m.ListingFacilities,
		m.Images,
		m.CostItems,
		m.IsOnline,
	)
}

func (m *HpdEntrustRelation) ValidateForCreate() error {
	if m == nil {
		return fmt.Errorf("hpd entrust relation is nil")
	}
	if m.ListingID.IsZero() {
		return fmt.Errorf("listingID is required")
	}
	if m.RelationStatus == HpdRelationStatusUnspecified || !m.RelationStatus.Valid() {
		return fmt.Errorf("relationStatus is invalid")
	}
	if commonmodel.IsBlank(m.OwnerPhone) && m.MaintainerStaffID.IsZero() && m.ServiceStaffID.IsZero() {
		return fmt.Errorf("ownerPhone or staffID is required")
	}
	if err := validateNonNegativeInt64("effectiveFrom", m.EffectiveFrom); err != nil {
		return err
	}
	return validateNonNegativeInt64("effectiveTo", m.EffectiveTo)
}

func ValidateHpdUpdateFields(fields bson.M) error {
	for key, value := range fields {
		if err := validateHpdUpdateField(key, value); err != nil {
			return err
		}
	}
	if err := validateHpdUpdateSourceAssetMode(fields); err != nil {
		return err
	}
	return nil
}

func validateHpdUpdateSourceAssetMode(fields bson.M) error {
	sourceValue, hasSourceType := fields["source_type"]
	assetValue, hasAssetMode := fields["asset_mode"]
	if !hasSourceType || !hasAssetMode {
		return nil
	}

	sourceType := HpdSourceType(fmt.Sprint(sourceValue))
	assetMode := HpdAssetMode(fmt.Sprint(assetValue))
	if !sourceType.Valid() || !assetMode.Valid() {
		return nil
	}
	if !HpdSourceAssetModeMatch(sourceType, assetMode) {
		return fmt.Errorf("sourceType and assetMode mismatch")
	}
	return nil
}

func validateHpdUpdateField(key string, value any) error {
	switch key {
	case "source_type":
		if !HpdSourceType(fmt.Sprint(value)).Valid() {
			return fmt.Errorf("sourceType is invalid")
		}
	case "asset_mode":
		if !HpdAssetMode(fmt.Sprint(value)).Valid() {
			return fmt.Errorf("assetMode is invalid")
		}
	case "listing_status":
		intValue, ok := hpdAnyInt(value)
		if !ok || intValue == int(HpdListingStatusUnspecified) || !HpdListingStatus(intValue).Valid() {
			return fmt.Errorf("listingStatus is invalid")
		}
	case "rent_mode":
		if !hmdmodel.RentMode(fmt.Sprint(value)).Valid() {
			return fmt.Errorf("rentMode is invalid")
		}
	case "city", "title":
		if text, ok := value.(string); !ok || commonmodel.IsBlank(text) {
			return fmt.Errorf("%s is required", key)
		}
	case "price", "subway_distance_m", "area_size", "weight_score":
		intValue, ok := hpdAnyInt(value)
		if !ok {
			return fmt.Errorf("%s must be int", key)
		}
		return validateNonNegativeInt(key, intValue)
	case "published_at", "offline_at":
		intValue, ok := hpdAnyInt64(value)
		if !ok {
			return fmt.Errorf("%s must be int64", key)
		}
		return validateNonNegativeInt64(key, intValue)
	case "orientation":
		if !hmdmodel.Orientation(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("orientation is invalid")
		}
	case "payment_cycle":
		if !hmdmodel.PaymentCycle(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("paymentCycle is invalid")
		}
	case "start_rent_rule":
		if !hmdmodel.StartRentRule(fmt.Sprint(value)).ValidOptional() {
			return fmt.Errorf("startRentRule is invalid")
		}
	case "listing_facilities":
		facilities, ok := value.([]hmdmodel.ListingFacility)
		if !ok {
			return fmt.Errorf("listingFacilities must be []ListingFacility")
		}
		return validateHpdListingFacilities(facilities)
	case "images":
		images, ok := value.([]hmdmodel.TaggedImage)
		if !ok {
			return fmt.Errorf("images must be []TaggedImage")
		}
		return validateHpdTaggedImages(images)
	case "cost_items":
		costItems, ok := value.([]HpdCostItem)
		if !ok {
			return fmt.Errorf("costItems must be []HpdCostItem")
		}
		return validateHpdCostItems(costItems)
	case "is_online":
		intValue, ok := hpdAnyInt(value)
		if !ok || !HpdOnlineStatus(intValue).Valid() {
			return fmt.Errorf("isOnline is invalid")
		}
	}
	return nil
}

func validateHpdMiniappListingFields(
	price int,
	subwayDistanceM int,
	areaSize int,
	weightScore int,
	orientation hmdmodel.Orientation,
	paymentCycle hmdmodel.PaymentCycle,
	startRentRule hmdmodel.StartRentRule,
	listingFacilities []hmdmodel.ListingFacility,
	images []hmdmodel.TaggedImage,
	costItems []HpdCostItem,
	isOnline HpdOnlineStatus,
) error {
	for name, value := range map[string]int{
		"price":           price,
		"subwayDistanceM": subwayDistanceM,
		"areaSize":        areaSize,
		"weightScore":     weightScore,
	} {
		if err := validateNonNegativeInt(name, value); err != nil {
			return err
		}
	}
	if !orientation.ValidOptional() {
		return fmt.Errorf("orientation is invalid")
	}
	if !paymentCycle.ValidOptional() {
		return fmt.Errorf("paymentCycle is invalid")
	}
	if !startRentRule.ValidOptional() {
		return fmt.Errorf("startRentRule is invalid")
	}
	if err := validateHpdListingFacilities(listingFacilities); err != nil {
		return err
	}
	if err := validateHpdTaggedImages(images); err != nil {
		return err
	}
	if err := validateHpdCostItems(costItems); err != nil {
		return err
	}
	if !isOnline.Valid() {
		return fmt.Errorf("isOnline is invalid")
	}
	return nil
}

func validateHpdListingFacilities(facilities []hmdmodel.ListingFacility) error {
	for _, facility := range facilities {
		if !facility.Valid() {
			return fmt.Errorf("listing facility %q is invalid", facility)
		}
	}
	return nil
}

func validateHpdTaggedImages(images []hmdmodel.TaggedImage) error {
	for _, image := range images {
		if !image.Tag.ValidOptional() {
			return fmt.Errorf("image tag %q is invalid", image.Tag)
		}
	}
	return nil
}

func validateHpdCostItems(costItems []HpdCostItem) error {
	for _, item := range costItems {
		if item.Amount < 0 {
			return fmt.Errorf("cost item amount must be non-negative")
		}
	}
	return nil
}

func validateNonNegativeInt(name string, value int) error {
	if value < 0 {
		return fmt.Errorf("%s must be non-negative", name)
	}
	return nil
}

func validateNonNegativeInt64(name string, value int64) error {
	if value < 0 {
		return fmt.Errorf("%s must be non-negative", name)
	}
	return nil
}

func hpdAnyInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case HpdListingStatus:
		return int(v), true
	case HpdOnlineStatus:
		return int(v), true
	default:
		return 0, false
	}
}

func hpdAnyInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	default:
		return 0, false
	}
}
