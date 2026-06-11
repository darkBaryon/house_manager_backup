package hmd

import (
	"fmt"
	"strings"

	hmdmodel "house-manager/internal/model/hmd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type BuildingBaseInfoUpdate struct {
	BuildingName      string
	FloorTotal        int
	ManagerName       string
	ManagerPhone      string
	Photos            []string
	ListingFacilities []hmdmodel.ListingFacility
}

func buildingBaseInfoUpdateFields(update BuildingBaseInfoUpdate) (bson.M, error) {
	if err := validateRequiredText("building_name", update.BuildingName); err != nil {
		return nil, err
	}
	if err := validateNonNegativeInt("floor_total", update.FloorTotal); err != nil {
		return nil, err
	}
	if err := validateListingFacilities(update.ListingFacilities); err != nil {
		return nil, err
	}
	return bson.M{
		"building_name":      update.BuildingName,
		"floor_total":        update.FloorTotal,
		"manager_name":       update.ManagerName,
		"manager_phone":      update.ManagerPhone,
		"photos":             update.Photos,
		"listing_facilities": update.ListingFacilities,
	}, nil
}

func validateRequiredText(name string, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s 不能为空", name)
	}
	return nil
}

func validateNonNegativeInt(name string, value int) error {
	if value < 0 {
		return fmt.Errorf("%s 不能小于 0", name)
	}
	return nil
}

func validateListingFacilities(facilities []hmdmodel.ListingFacility) error {
	for _, facility := range facilities {
		if !facility.Valid() {
			return fmt.Errorf("房源设施 %q 不合法", facility)
		}
	}
	return nil
}
