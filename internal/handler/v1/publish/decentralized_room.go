package publish

import (
	publishsvc "house-manager/internal/service/publish"

	"github.com/gin-gonic/gin"
)

func (h *PublishHandler) CreateDecentralizedRoom(c *gin.Context) {
	var req decentralizedRoomRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	decentralizedID, ok := objectIDFromHex(c, req.DecentralizedID)
	if !ok {
		return
	}
	result, err := h.service.CreateDecentralizedRoom(c.Request.Context(), req.toCreateInput(decentralizedID))
	writePublishResult(c, "create decentralized room failed", result, err)
}

func (h *PublishHandler) DecentralizedRoomDetail(c *gin.Context) {
	id, ok := bindIDRequest(c)
	if !ok {
		return
	}
	result, err := h.service.GetDecentralizedRoom(c.Request.Context(), id)
	writePublishResult(c, "get decentralized room failed", result, err)
}

func (h *PublishHandler) ListDecentralizedRoomsByCommunity(c *gin.Context) {
	var req decentralizedIDRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	decentralizedID, ok := objectIDFromHex(c, req.DecentralizedID)
	if !ok {
		return
	}
	result, err := h.service.ListDecentralizedRoomsByCommunity(c.Request.Context(), decentralizedID)
	writePublishResult(c, "list decentralized rooms failed", result, err)
}

func (h *PublishHandler) UpdateDecentralizedRoom(c *gin.Context) {
	var req decentralizedRoomRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	id, ok := objectIDFromHex(c, req.ID)
	if !ok {
		return
	}
	result, err := h.service.UpdateDecentralizedRoom(c.Request.Context(), req.toUpdateInput(id))
	writePublishResult(c, "update decentralized room failed", result, err)
}

func (h *PublishHandler) UpdateDecentralizedRoomStatus(c *gin.Context) {
	id, roomStatus, ok := bindRoomStatusRequest(c)
	if !ok {
		return
	}
	result, err := h.service.UpdateDecentralizedRoomStatus(c.Request.Context(), publishsvc.UpdateDecentralizedRoomStatusInput{
		ID:         id,
		RoomStatus: roomStatus,
	})
	writePublishResult(c, "update decentralized room status failed", result, err)
}
