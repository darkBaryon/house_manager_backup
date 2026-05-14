package community

import (
	"house-manager/internal/handler/v1/publish/common"
	"house-manager/internal/model"
)

type request struct {
	ID            string                  `json:"id"`
	CommunityName string                  `json:"community_name" binding:"required"`
	City          string                  `json:"city" binding:"required"`
	District      string                  `json:"district"`
	BizArea       string                  `json:"biz_area"`
	AddressText   string                  `json:"address_text"`
	Geo           *common.GeoPointRequest `json:"geo"`
	SubwayStation string                  `json:"subway_station"`
}

type response struct {
	common.EntityMetaResponse

	CommunityName string                   `json:"community_name"`
	City          string                   `json:"city"`
	District      string                   `json:"district"`
	BizArea       string                   `json:"biz_area"`
	AddressText   string                   `json:"address_text"`
	Geo           *common.GeoPointResponse `json:"geo"`
	SubwayStation string                   `json:"subway_station"`
}

func toResponse(community *model.HmdDecentralized) *response {
	if community == nil {
		return nil
	}
	return &response{
		EntityMetaResponse: common.EntityMeta(community.CommonFields),
		CommunityName:      community.CommunityName,
		City:               community.City,
		District:           community.District,
		BizArea:            community.BizArea,
		AddressText:        community.AddressText,
		Geo:                common.GeoPoint(community.Geo),
		SubwayStation:      community.SubwayStation,
	}
}

func toListResponse(communities []model.HmdDecentralized) common.ListResponse[response] {
	list := make([]response, 0, len(communities))
	for i := range communities {
		item := toResponse(&communities[i])
		if item != nil {
			list = append(list, *item)
		}
	}
	return common.ListResponse[response]{List: list}
}
