package project

import (
	"context"

	"house-manager/internal/handler/v1/publish/common"
	"house-manager/internal/model"
	publishsvc "house-manager/internal/service/publish"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service interface {
	CreateCentralizedProject(ctx context.Context, input publishsvc.CreateCentralizedProjectInput) (*model.HmdCentralized, error)
	GetCentralizedProject(ctx context.Context, id bson.ObjectID) (*model.HmdCentralized, error)
	ListCentralizedProjects(ctx context.Context, input publishsvc.ListCentralizedProjectsInput) ([]model.HmdCentralized, error)
	UpdateCentralizedProject(ctx context.Context, input publishsvc.UpdateCentralizedProjectInput) (*model.HmdCentralized, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/centralized_project/create", h.Create)
	rg.POST("/centralized_project/detail", h.Detail)
	rg.POST("/centralized_project/update", h.Update)
	rg.POST("/centralized_project/list", h.List)
}

func (h *Handler) Create(c *gin.Context) {
	var req createRequest
	if !common.BindJSON(c, &req) {
		return
	}
	result, err := h.service.CreateCentralizedProject(c.Request.Context(), publishsvc.CreateCentralizedProjectInput{
		ProjectName: req.ProjectName,
		ProjectCode: req.ProjectCode,
		City:        req.City,
		District:    req.District,
		AddressText: req.AddressText,
		Geo:         req.Geo.ToServiceInput(),
		BrandName:   req.BrandName,
	})
	common.WriteResult(c, "create centralized project failed", toResponse(result), err)
}

func (h *Handler) Detail(c *gin.Context) {
	id, ok := common.BindID(c)
	if !ok {
		return
	}
	result, err := h.service.GetCentralizedProject(c.Request.Context(), id)
	common.WriteResult(c, "get centralized project failed", toResponse(result), err)
}

func (h *Handler) List(c *gin.Context) {
	var req listRequest
	if !common.BindJSON(c, &req) {
		return
	}
	result, err := h.service.ListCentralizedProjects(c.Request.Context(), publishsvc.ListCentralizedProjectsInput{
		City:     req.City,
		District: req.District,
	})
	common.WriteResult(c, "list centralized projects failed", toListResponse(result), err)
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
	result, err := h.service.UpdateCentralizedProject(c.Request.Context(), publishsvc.UpdateCentralizedProjectInput{
		ID:          id,
		ProjectName: req.ProjectName,
		City:        req.City,
		District:    req.District,
		AddressText: req.AddressText,
		Geo:         req.Geo.ToServiceInput(),
		BrandName:   req.BrandName,
	})
	common.WriteResult(c, "update centralized project failed", toResponse(result), err)
}
