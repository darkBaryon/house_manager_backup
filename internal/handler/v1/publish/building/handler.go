package building

import (
	"context"

	"house-manager/internal/handler/v1/publish/common"
	"house-manager/internal/model"
	publishsvc "house-manager/internal/service/publish"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service interface {
	CreateBuilding(ctx context.Context, input publishsvc.CreateBuildingInput) (*model.HmdBuilding, error)
	GetBuilding(ctx context.Context, id bson.ObjectID) (*model.HmdBuilding, error)
	ListBuildingsByProject(ctx context.Context, projectID bson.ObjectID) ([]model.HmdBuilding, error)
	UpdateBuilding(ctx context.Context, input publishsvc.UpdateBuildingInput) (*model.HmdBuilding, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/building/create", h.Create)
	rg.POST("/building/detail", h.Detail)
	rg.POST("/building/update", h.Update)
	rg.POST("/building/list_by_project", h.ListByProject)
}

func (h *Handler) Create(c *gin.Context) {
	var req createRequest
	if !common.BindJSON(c, &req) {
		return
	}
	projectID, ok := common.ObjectIDFromHex(c, req.ProjectID)
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
	common.WriteResult(c, "create building failed", toResponse(result), err)
}

func (h *Handler) Detail(c *gin.Context) {
	id, ok := common.BindID(c)
	if !ok {
		return
	}
	result, err := h.service.GetBuilding(c.Request.Context(), id)
	common.WriteResult(c, "get building failed", toResponse(result), err)
}

func (h *Handler) ListByProject(c *gin.Context) {
	projectID, ok := common.BindProjectID(c)
	if !ok {
		return
	}
	result, err := h.service.ListBuildingsByProject(c.Request.Context(), projectID)
	common.WriteResult(c, "list buildings failed", toListResponse(result), err)
}

func (h *Handler) Update(c *gin.Context) {
	var req updateRequest
	if !common.BindJSON(c, &req) {
		return
	}
	id, ok := common.ObjectIDFromHex(c, req.ID)
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
	common.WriteResult(c, "update building failed", toResponse(result), err)
}
