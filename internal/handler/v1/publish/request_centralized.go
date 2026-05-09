package publish

type createCentralizedProjectRequest struct {
	ProjectName string           `json:"project_name" binding:"required"`
	ProjectCode string           `json:"project_code" binding:"required"`
	City        string           `json:"city" binding:"required"`
	District    string           `json:"district"`
	AddressText string           `json:"address_text"`
	Geo         *geoPointRequest `json:"geo"`
	BrandName   string           `json:"brand_name"`
}

type updateCentralizedProjectRequest struct {
	ID          string           `json:"id" binding:"required"`
	ProjectName string           `json:"project_name" binding:"required"`
	City        string           `json:"city" binding:"required"`
	District    string           `json:"district"`
	AddressText string           `json:"address_text"`
	Geo         *geoPointRequest `json:"geo"`
	BrandName   string           `json:"brand_name"`
}
