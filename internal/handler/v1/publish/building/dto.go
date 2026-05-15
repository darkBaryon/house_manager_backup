package building

import (
	"house-manager/internal/handler/v1/publish/common"
	hmdmodel "house-manager/internal/model/hmd"
)

type createRequest struct {
	ProjectID         string   `json:"project_id" binding:"required"`
	BuildingName      string   `json:"building_name" binding:"required"`
	BuildingCode      string   `json:"building_code"`
	FloorTotal        int      `json:"floor_total"`
	ManagerName       string   `json:"manager_name"`
	ManagerPhone      string   `json:"manager_phone"`
	Photos            []string `json:"photos"`
	ListingFacilities []string `json:"listing_facilities"`
}

type updateRequest struct {
	ID                string   `json:"id" binding:"required"`
	BuildingName      string   `json:"building_name" binding:"required"`
	FloorTotal        int      `json:"floor_total"`
	ManagerName       string   `json:"manager_name"`
	ManagerPhone      string   `json:"manager_phone"`
	Photos            []string `json:"photos"`
	ListingFacilities []string `json:"listing_facilities"`
}

type response struct {
	common.EntityMetaResponse

	ProjectID         string   `json:"project_id"`
	BuildingName      string   `json:"building_name"`
	BuildingCode      string   `json:"building_code"`
	FloorTotal        int      `json:"floor_total"`
	ManagerName       string   `json:"manager_name"`
	ManagerPhone      string   `json:"manager_phone"`
	Photos            []string `json:"photos"`
	ListingFacilities []string `json:"listing_facilities"`
}

func toResponse(building *hmdmodel.HmdBuilding) *response {
	if building == nil {
		return nil
	}
	return &response{
		EntityMetaResponse: common.EntityMeta(building.CommonFields),
		ProjectID:          common.ObjectIDHex(building.ProjectID),
		BuildingName:       building.BuildingName,
		BuildingCode:       building.BuildingCode,
		FloorTotal:         building.FloorTotal,
		ManagerName:        building.ManagerName,
		ManagerPhone:       building.ManagerPhone,
		Photos:             common.StringsOrEmpty(building.Photos),
		ListingFacilities:  common.ListingFacilities(building.ListingFacilities),
	}
}

func toListResponse(buildings []hmdmodel.HmdBuilding) common.ListResponse[response] {
	list := make([]response, 0, len(buildings))
	for i := range buildings {
		item := toResponse(&buildings[i])
		if item != nil {
			list = append(list, *item)
		}
	}
	return common.ListResponse[response]{List: list}
}
