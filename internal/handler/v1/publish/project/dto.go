package project

import (
	"house-manager/internal/handler/v1/publish/common"
	hmdmodel "house-manager/internal/model/hmd"
)

type createRequest struct {
	ProjectName string                  `json:"project_name" binding:"required"`
	ProjectCode string                  `json:"project_code" binding:"required"`
	City        string                  `json:"city" binding:"required"`
	District    string                  `json:"district"`
	AddressText string                  `json:"address_text"`
	Geo         *common.GeoPointRequest `json:"geo"`
	BrandName   string                  `json:"brand_name"`
}

type updateRequest struct {
	ID          string                  `json:"id" binding:"required"`
	ProjectName string                  `json:"project_name" binding:"required"`
	City        string                  `json:"city" binding:"required"`
	District    string                  `json:"district"`
	AddressText string                  `json:"address_text"`
	Geo         *common.GeoPointRequest `json:"geo"`
	BrandName   string                  `json:"brand_name"`
}

type listRequest struct {
	City     string `json:"city"`
	District string `json:"district"`
}

type response struct {
	common.EntityMetaResponse

	ProjectName string                   `json:"project_name"`
	ProjectCode string                   `json:"project_code"`
	City        string                   `json:"city"`
	District    string                   `json:"district"`
	AddressText string                   `json:"address_text"`
	Geo         *common.GeoPointResponse `json:"geo"`
	BrandName   string                   `json:"brand_name"`
}

func toResponse(project *hmdmodel.HmdCentralized) *response {
	if project == nil {
		return nil
	}
	return &response{
		EntityMetaResponse: common.EntityMeta(project.CommonFields),
		ProjectName:        project.ProjectName,
		ProjectCode:        project.ProjectCode,
		City:               project.City,
		District:           project.District,
		AddressText:        project.AddressText,
		Geo:                common.GeoPoint(project.Geo),
		BrandName:          project.BrandName,
	}
}

func toListResponse(projects []hmdmodel.HmdCentralized) common.ListResponse[response] {
	list := make([]response, 0, len(projects))
	for i := range projects {
		item := toResponse(&projects[i])
		if item != nil {
			list = append(list, *item)
		}
	}
	return common.ListResponse[response]{List: list}
}
