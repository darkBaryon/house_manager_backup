package publish

type createBuildingRequest struct {
	ProjectID         string   `json:"project_id" binding:"required"`
	BuildingName      string   `json:"building_name" binding:"required"`
	BuildingCode      string   `json:"building_code"`
	FloorTotal        int      `json:"floor_total"`
	ManagerName       string   `json:"manager_name"`
	ManagerPhone      string   `json:"manager_phone"`
	Photos            []string `json:"photos"`
	ListingFacilities []string `json:"listing_facilities"`
}

type updateBuildingRequest struct {
	ID                string   `json:"id" binding:"required"`
	BuildingName      string   `json:"building_name" binding:"required"`
	FloorTotal        int      `json:"floor_total"`
	ManagerName       string   `json:"manager_name"`
	ManagerPhone      string   `json:"manager_phone"`
	Photos            []string `json:"photos"`
	ListingFacilities []string `json:"listing_facilities"`
}
