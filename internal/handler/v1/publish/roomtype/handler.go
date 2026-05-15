package roomtype

import (
	"context"

	"house-manager/internal/handler/v1/publish/common"
	hmdmodel "house-manager/internal/model/hmd"
	publishsvc "house-manager/internal/service/publish"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service interface {
	CreateRoomType(ctx context.Context, input publishsvc.CreateRoomTypeInput) (*hmdmodel.HmdRoomTypeCentralized, error)
	GetRoomType(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdRoomTypeCentralized, error)
	ListRoomTypesByProject(ctx context.Context, projectID bson.ObjectID) ([]hmdmodel.HmdRoomTypeCentralized, error)
	ListRoomTypesByBuilding(ctx context.Context, buildingID bson.ObjectID) ([]hmdmodel.HmdRoomTypeCentralized, error)
	UpdateRoomType(ctx context.Context, input publishsvc.UpdateRoomTypeInput) (*hmdmodel.HmdRoomTypeCentralized, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/room_type/create", h.Create)
	rg.POST("/room_type/detail", h.Detail)
	rg.POST("/room_type/update", h.Update)
	rg.POST("/room_type/list_by_project", h.ListByProject)
	rg.POST("/room_type/list_by_building", h.ListByBuilding)
}

func (h *Handler) Create(c *gin.Context) {
	var req request
	if !common.BindJSON(c, &req) {
		return
	}
	projectID, ok := common.OptionalObjectIDFromHex(c, req.ProjectID)
	if !ok {
		return
	}
	buildingID, ok := common.OptionalObjectIDFromHex(c, req.BuildingID)
	if !ok {
		return
	}
	result, err := h.service.CreateRoomType(c.Request.Context(), req.toCreateInput(projectID, buildingID))
	common.WriteResult(c, "create room type failed", toResponse(result), err)
}

func (h *Handler) Detail(c *gin.Context) {
	id, ok := common.BindID(c)
	if !ok {
		return
	}
	result, err := h.service.GetRoomType(c.Request.Context(), id)
	common.WriteResult(c, "get room type failed", toResponse(result), err)
}

func (h *Handler) ListByProject(c *gin.Context) {
	projectID, ok := common.BindProjectID(c)
	if !ok {
		return
	}
	result, err := h.service.ListRoomTypesByProject(c.Request.Context(), projectID)
	common.WriteResult(c, "list room types by project failed", toListResponse(result), err)
}

func (h *Handler) ListByBuilding(c *gin.Context) {
	buildingID, ok := common.BindBuildingID(c)
	if !ok {
		return
	}
	result, err := h.service.ListRoomTypesByBuilding(c.Request.Context(), buildingID)
	common.WriteResult(c, "list room types by building failed", toListResponse(result), err)
}

func (h *Handler) Update(c *gin.Context) {
	var req request
	if !common.BindJSON(c, &req) {
		return
	}
	id, ok := common.ObjectIDFromHex(c, req.ID)
	if !ok {
		return
	}
	result, err := h.service.UpdateRoomType(c.Request.Context(), req.toUpdateInput(id))
	common.WriteResult(c, "update room type failed", toResponse(result), err)
}
