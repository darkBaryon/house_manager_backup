package publish

import (
	publishsvc "house-manager/internal/service/publish"

	"github.com/gin-gonic/gin"
)

func (h *PublishHandler) CreateDecentralizedCommunity(c *gin.Context) {
	var req decentralizedCommunityRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	result, err := h.service.CreateDecentralizedCommunity(c.Request.Context(), publishsvc.CreateDecentralizedCommunityInput{
		CommunityName: req.CommunityName,
		City:          req.City,
		District:      req.District,
		BizArea:       req.BizArea,
		AddressText:   req.AddressText,
		Geo:           req.Geo.toServiceInput(),
		SubwayStation: req.SubwayStation,
	})
	writePublishResult(c, "create decentralized community failed", result, err)
}

func (h *PublishHandler) DecentralizedCommunityDetail(c *gin.Context) {
	id, ok := bindIDRequest(c)
	if !ok {
		return
	}
	result, err := h.service.GetDecentralizedCommunity(c.Request.Context(), id)
	writePublishResult(c, "get decentralized community failed", result, err)
}

func (h *PublishHandler) ListDecentralizedCommunities(c *gin.Context) {
	var req listByCityRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	result, err := h.service.ListDecentralizedCommunities(c.Request.Context(), publishsvc.ListDecentralizedCommunitiesInput{
		City:     req.City,
		District: req.District,
	})
	writePublishResult(c, "list decentralized communities failed", result, err)
}

func (h *PublishHandler) UpdateDecentralizedCommunity(c *gin.Context) {
	var req decentralizedCommunityRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	id, ok := objectIDFromHex(c, req.ID)
	if !ok {
		return
	}
	result, err := h.service.UpdateDecentralizedCommunity(c.Request.Context(), publishsvc.UpdateDecentralizedCommunityInput{
		ID:            id,
		CommunityName: req.CommunityName,
		City:          req.City,
		District:      req.District,
		BizArea:       req.BizArea,
		AddressText:   req.AddressText,
		Geo:           req.Geo.toServiceInput(),
		SubwayStation: req.SubwayStation,
	})
	writePublishResult(c, "update decentralized community failed", result, err)
}
