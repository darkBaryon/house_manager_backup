package publish

import (
	publishsvc "house-manager/internal/service/publish"

	"github.com/gin-gonic/gin"
)

func (h *PublishHandler) CreateCentralizedRoom(c *gin.Context) {
	var req centralizedRoomRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	projectID, buildingID, roomTypeID, ok := parseCentralizedRoomIDs(c, req.ProjectID, req.BuildingID, req.RoomTypeID)
	if !ok {
		return
	}
	result, err := h.service.CreateCentralizedRoom(c.Request.Context(), req.toCreateInput(projectID, buildingID, roomTypeID))
	writePublishResult(c, "create centralized room failed", result, err)
}

func (h *PublishHandler) CentralizedRoomDetail(c *gin.Context) {
	id, ok := bindIDRequest(c)
	if !ok {
		return
	}
	result, err := h.service.GetCentralizedRoom(c.Request.Context(), id)
	writePublishResult(c, "get centralized room failed", result, err)
}

func (h *PublishHandler) ListCentralizedRoomsByProject(c *gin.Context) {
	projectID, ok := bindProjectIDRequest(c)
	if !ok {
		return
	}
	result, err := h.service.ListCentralizedRoomsByProject(c.Request.Context(), projectID)
	writePublishResult(c, "list centralized rooms by project failed", result, err)
}

func (h *PublishHandler) ListCentralizedRoomsByBuilding(c *gin.Context) {
	buildingID, ok := bindBuildingIDRequest(c)
	if !ok {
		return
	}
	result, err := h.service.ListCentralizedRoomsByBuilding(c.Request.Context(), buildingID)
	writePublishResult(c, "list centralized rooms by building failed", result, err)
}

func (h *PublishHandler) UpdateCentralizedRoom(c *gin.Context) {
	var req centralizedRoomRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	id, ok := objectIDFromHex(c, req.ID)
	if !ok {
		return
	}
	result, err := h.service.UpdateCentralizedRoom(c.Request.Context(), req.toUpdateInput(id))
	writePublishResult(c, "update centralized room failed", result, err)
}

func (h *PublishHandler) UpdateCentralizedRoomStatus(c *gin.Context) {
	id, roomStatus, ok := bindRoomStatusRequest(c)
	if !ok {
		return
	}
	result, err := h.service.UpdateCentralizedRoomStatus(c.Request.Context(), publishsvc.UpdateCentralizedRoomStatusInput{
		ID:         id,
		RoomStatus: roomStatus,
	})
	writePublishResult(c, "update centralized room status failed", result, err)
}
