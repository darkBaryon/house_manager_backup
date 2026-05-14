package room

import (
	"context"

	"house-manager/internal/handler/v1/publish/common"
	"house-manager/internal/model"
	publishsvc "house-manager/internal/service/publish"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service interface {
	CreateCentralizedRoom(ctx context.Context, input publishsvc.CreateCentralizedRoomInput) (*model.HmdRoomCentralized, error)
	GetCentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomCentralized, error)
	ListCentralizedRoomsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdRoomCentralized, error)
	ListCentralizedRoomsByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]model.HmdRoomCentralized, error)
	UpdateCentralizedRoom(ctx context.Context, input publishsvc.UpdateCentralizedRoomInput) (*model.HmdRoomCentralized, error)
	UpdateCentralizedRoomStatus(ctx context.Context, input publishsvc.UpdateCentralizedRoomStatusInput) (*model.HmdRoomCentralized, error)
	CreateDecentralizedRoom(ctx context.Context, input publishsvc.CreateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error)
	GetDecentralizedRoom(ctx context.Context, id bson.ObjectID) (*model.HmdRoomDecentralized, error)
	ListDecentralizedRoomsByCommunity(ctx context.Context, decentralizedID bson.ObjectID) ([]model.HmdRoomDecentralized, error)
	UpdateDecentralizedRoom(ctx context.Context, input publishsvc.UpdateDecentralizedRoomInput) (*model.HmdRoomDecentralized, error)
	UpdateDecentralizedRoomStatus(ctx context.Context, input publishsvc.UpdateDecentralizedRoomStatusInput) (*model.HmdRoomDecentralized, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/centralized_room/create", h.CreateCentralized)
	rg.POST("/centralized_room/detail", h.CentralizedDetail)
	rg.POST("/centralized_room/update", h.UpdateCentralized)
	rg.POST("/centralized_room/update_status", h.UpdateCentralizedStatus)
	rg.POST("/centralized_room/list_by_project", h.ListCentralizedByProject)
	rg.POST("/centralized_room/list_by_building", h.ListCentralizedByBuilding)

	rg.POST("/decentralized_room/create", h.CreateDecentralized)
	rg.POST("/decentralized_room/detail", h.DecentralizedDetail)
	rg.POST("/decentralized_room/update", h.UpdateDecentralized)
	rg.POST("/decentralized_room/update_status", h.UpdateDecentralizedStatus)
	rg.POST("/decentralized_room/list_by_community", h.ListDecentralizedByCommunity)
}

func (h *Handler) CreateCentralized(c *gin.Context) {
	var req centralizedRequest
	if !common.BindJSON(c, &req) {
		return
	}
	projectID, buildingID, roomTypeID, ok := parseCentralizedIDs(c, req.ProjectID, req.BuildingID, req.RoomTypeID)
	if !ok {
		return
	}
	result, err := h.service.CreateCentralizedRoom(c.Request.Context(), req.toCreateInput(projectID, buildingID, roomTypeID))
	common.WriteResult(c, "create centralized room failed", toCentralizedResponse(result), err)
}

func (h *Handler) CentralizedDetail(c *gin.Context) {
	id, ok := common.BindID(c)
	if !ok {
		return
	}
	result, err := h.service.GetCentralizedRoom(c.Request.Context(), id)
	common.WriteResult(c, "get centralized room failed", toCentralizedResponse(result), err)
}

func (h *Handler) ListCentralizedByProject(c *gin.Context) {
	projectID, ok := common.BindProjectID(c)
	if !ok {
		return
	}
	result, err := h.service.ListCentralizedRoomsByProject(c.Request.Context(), projectID)
	common.WriteResult(c, "list centralized rooms by project failed", toCentralizedListResponse(result), err)
}

func (h *Handler) ListCentralizedByBuilding(c *gin.Context) {
	buildingID, ok := common.BindBuildingID(c)
	if !ok {
		return
	}
	result, err := h.service.ListCentralizedRoomsByBuilding(c.Request.Context(), buildingID)
	common.WriteResult(c, "list centralized rooms by building failed", toCentralizedListResponse(result), err)
}

func (h *Handler) UpdateCentralized(c *gin.Context) {
	var req centralizedRequest
	if !common.BindJSON(c, &req) {
		return
	}
	id, ok := common.ObjectIDFromHex(c, req.ID)
	if !ok {
		return
	}
	result, err := h.service.UpdateCentralizedRoom(c.Request.Context(), req.toUpdateInput(id))
	common.WriteResult(c, "update centralized room failed", toCentralizedResponse(result), err)
}

func (h *Handler) UpdateCentralizedStatus(c *gin.Context) {
	id, roomStatus, ok := common.BindRoomStatus(c)
	if !ok {
		return
	}
	result, err := h.service.UpdateCentralizedRoomStatus(c.Request.Context(), publishsvc.UpdateCentralizedRoomStatusInput{
		ID:         id,
		RoomStatus: roomStatus,
	})
	common.WriteResult(c, "update centralized room status failed", toCentralizedResponse(result), err)
}

func (h *Handler) CreateDecentralized(c *gin.Context) {
	var req decentralizedRequest
	if !common.BindJSON(c, &req) {
		return
	}
	decentralizedID, ok := common.ObjectIDFromHex(c, req.DecentralizedID)
	if !ok {
		return
	}
	result, err := h.service.CreateDecentralizedRoom(c.Request.Context(), req.toCreateInput(decentralizedID))
	common.WriteResult(c, "create decentralized room failed", toDecentralizedResponse(result), err)
}

func (h *Handler) DecentralizedDetail(c *gin.Context) {
	id, ok := common.BindID(c)
	if !ok {
		return
	}
	result, err := h.service.GetDecentralizedRoom(c.Request.Context(), id)
	common.WriteResult(c, "get decentralized room failed", toDecentralizedResponse(result), err)
}

func (h *Handler) ListDecentralizedByCommunity(c *gin.Context) {
	var req common.DecentralizedIDRequest
	if !common.BindJSON(c, &req) {
		return
	}
	decentralizedID, ok := common.ObjectIDFromHex(c, req.DecentralizedID)
	if !ok {
		return
	}
	result, err := h.service.ListDecentralizedRoomsByCommunity(c.Request.Context(), decentralizedID)
	common.WriteResult(c, "list decentralized rooms failed", toDecentralizedListResponse(result), err)
}

func (h *Handler) UpdateDecentralized(c *gin.Context) {
	var req decentralizedRequest
	if !common.BindJSON(c, &req) {
		return
	}
	id, ok := common.ObjectIDFromHex(c, req.ID)
	if !ok {
		return
	}
	result, err := h.service.UpdateDecentralizedRoom(c.Request.Context(), req.toUpdateInput(id))
	common.WriteResult(c, "update decentralized room failed", toDecentralizedResponse(result), err)
}

func (h *Handler) UpdateDecentralizedStatus(c *gin.Context) {
	id, roomStatus, ok := common.BindRoomStatus(c)
	if !ok {
		return
	}
	result, err := h.service.UpdateDecentralizedRoomStatus(c.Request.Context(), publishsvc.UpdateDecentralizedRoomStatusInput{
		ID:         id,
		RoomStatus: roomStatus,
	})
	common.WriteResult(c, "update decentralized room status failed", toDecentralizedResponse(result), err)
}

func parseCentralizedIDs(c *gin.Context, projectIDText, buildingIDText, roomTypeIDText string) (bson.ObjectID, bson.ObjectID, bson.ObjectID, bool) {
	projectID, ok := common.ObjectIDFromHex(c, projectIDText)
	if !ok {
		return bson.NilObjectID, bson.NilObjectID, bson.NilObjectID, false
	}
	buildingID, ok := common.ObjectIDFromHex(c, buildingIDText)
	if !ok {
		return bson.NilObjectID, bson.NilObjectID, bson.NilObjectID, false
	}
	roomTypeID, ok := common.OptionalObjectIDFromHex(c, roomTypeIDText)
	if !ok {
		return bson.NilObjectID, bson.NilObjectID, bson.NilObjectID, false
	}
	return projectID, buildingID, roomTypeID, true
}
