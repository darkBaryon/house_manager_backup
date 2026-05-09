package publish

import (
	publishsvc "house-manager/internal/service/publish"

	"github.com/gin-gonic/gin"
)

func (h *PublishHandler) CreateBuilding(c *gin.Context) {
	var req createBuildingRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	projectID, ok := objectIDFromHex(c, req.ProjectID)
	if !ok {
		return
	}
	result, err := h.service.CreateBuilding(c.Request.Context(), publishsvc.CreateBuildingInput{
		ProjectID:         projectID,
		BuildingName:      req.BuildingName,
		BuildingCode:      req.BuildingCode,
		FloorTotal:        req.FloorTotal,
		ManagerName:       req.ManagerName,
		ManagerPhone:      req.ManagerPhone,
		Photos:            req.Photos,
		ListingFacilities: req.ListingFacilities,
	})
	writePublishResult(c, "create building failed", result, err)
}

func (h *PublishHandler) BuildingDetail(c *gin.Context) {
	id, ok := bindIDRequest(c)
	if !ok {
		return
	}
	result, err := h.service.GetBuilding(c.Request.Context(), id)
	writePublishResult(c, "get building failed", result, err)
}

func (h *PublishHandler) ListBuildingsByProject(c *gin.Context) {
	projectID, ok := bindProjectIDRequest(c)
	if !ok {
		return
	}
	result, err := h.service.ListBuildingsByProject(c.Request.Context(), projectID)
	writePublishResult(c, "list buildings failed", result, err)
}

func (h *PublishHandler) UpdateBuilding(c *gin.Context) {
	var req updateBuildingRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	id, ok := objectIDFromHex(c, req.ID)
	if !ok {
		return
	}
	result, err := h.service.UpdateBuilding(c.Request.Context(), publishsvc.UpdateBuildingInput{
		ID:                id,
		BuildingName:      req.BuildingName,
		FloorTotal:        req.FloorTotal,
		ManagerName:       req.ManagerName,
		ManagerPhone:      req.ManagerPhone,
		Photos:            req.Photos,
		ListingFacilities: req.ListingFacilities,
	})
	writePublishResult(c, "update building failed", result, err)
}
