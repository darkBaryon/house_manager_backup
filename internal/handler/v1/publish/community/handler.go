package community

import (
	"context"

	"house-manager/internal/handler/v1/publish/common"
	hmdmodel "house-manager/internal/model/hmd"
	publishsvc "house-manager/internal/service/publish"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service interface {
	CreateDecentralizedCommunity(ctx context.Context, input publishsvc.CreateDecentralizedCommunityInput) (*hmdmodel.HmdDecentralized, error)
	GetDecentralizedCommunity(ctx context.Context, id bson.ObjectID) (*hmdmodel.HmdDecentralized, error)
	ListDecentralizedCommunities(ctx context.Context, input publishsvc.ListDecentralizedCommunitiesInput) ([]hmdmodel.HmdDecentralized, error)
	UpdateDecentralizedCommunity(ctx context.Context, input publishsvc.UpdateDecentralizedCommunityInput) (*hmdmodel.HmdDecentralized, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/decentralized_community/create", h.Create)
	rg.POST("/decentralized_community/detail", h.Detail)
	rg.POST("/decentralized_community/update", h.Update)
	rg.POST("/decentralized_community/list", h.List)
}

func (h *Handler) Create(c *gin.Context) {
	var req request
	if !common.BindJSON(c, &req) {
		return
	}
	result, err := h.service.CreateDecentralizedCommunity(c.Request.Context(), publishsvc.CreateDecentralizedCommunityInput{
		CommunityName: req.CommunityName,
		City:          req.City,
		District:      req.District,
		BizArea:       req.BizArea,
		AddressText:   req.AddressText,
		Geo:           req.Geo.ToServiceInput(),
		SubwayStation: req.SubwayStation,
	})
	common.WriteResult(c, "create decentralized community failed", toResponse(result), err)
}

func (h *Handler) Detail(c *gin.Context) {
	id, ok := common.BindID(c)
	if !ok {
		return
	}
	result, err := h.service.GetDecentralizedCommunity(c.Request.Context(), id)
	common.WriteResult(c, "get decentralized community failed", toResponse(result), err)
}

func (h *Handler) List(c *gin.Context) {
	var req common.OptionalListByCityRequest
	if !common.BindJSON(c, &req) {
		return
	}
	result, err := h.service.ListDecentralizedCommunities(c.Request.Context(), publishsvc.ListDecentralizedCommunitiesInput{
		City:     req.City,
		District: req.District,
	})
	common.WriteResult(c, "list decentralized communities failed", toListResponse(result), err)
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
	result, err := h.service.UpdateDecentralizedCommunity(c.Request.Context(), publishsvc.UpdateDecentralizedCommunityInput{
		ID:            id,
		CommunityName: req.CommunityName,
		City:          req.City,
		District:      req.District,
		BizArea:       req.BizArea,
		AddressText:   req.AddressText,
		Geo:           req.Geo.ToServiceInput(),
		SubwayStation: req.SubwayStation,
	})
	common.WriteResult(c, "update decentralized community failed", toResponse(result), err)
}
