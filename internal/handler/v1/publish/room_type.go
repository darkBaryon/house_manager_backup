package publish

import "github.com/gin-gonic/gin"

func (h *PublishHandler) CreateRoomType(c *gin.Context) {
	var req roomTypeRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	projectID, ok := optionalObjectIDFromHex(c, req.ProjectID)
	if !ok {
		return
	}
	buildingID, ok := optionalObjectIDFromHex(c, req.BuildingID)
	if !ok {
		return
	}
	result, err := h.service.CreateRoomType(c.Request.Context(), req.toCreateInput(projectID, buildingID))
	writePublishResult(c, "create room type failed", result, err)
}

func (h *PublishHandler) RoomTypeDetail(c *gin.Context) {
	id, ok := bindIDRequest(c)
	if !ok {
		return
	}
	result, err := h.service.GetRoomType(c.Request.Context(), id)
	writePublishResult(c, "get room type failed", result, err)
}

func (h *PublishHandler) ListRoomTypesByProject(c *gin.Context) {
	projectID, ok := bindProjectIDRequest(c)
	if !ok {
		return
	}
	result, err := h.service.ListRoomTypesByProject(c.Request.Context(), projectID)
	writePublishResult(c, "list room types by project failed", result, err)
}

func (h *PublishHandler) ListRoomTypesByBuilding(c *gin.Context) {
	buildingID, ok := bindBuildingIDRequest(c)
	if !ok {
		return
	}
	result, err := h.service.ListRoomTypesByBuilding(c.Request.Context(), buildingID)
	writePublishResult(c, "list room types by building failed", result, err)
}

func (h *PublishHandler) UpdateRoomType(c *gin.Context) {
	var req roomTypeRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	id, ok := objectIDFromHex(c, req.ID)
	if !ok {
		return
	}
	result, err := h.service.UpdateRoomType(c.Request.Context(), req.toUpdateInput(id))
	writePublishResult(c, "update room type failed", result, err)
}
