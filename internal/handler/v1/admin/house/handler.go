package house

import (
	"context"
	"fmt"
	"io"

	"house-manager/internal/handler"
	"house-manager/internal/middleware"
	housesvc "house-manager/internal/service/admin/house"
	"house-manager/pkg/errcode"
	"house-manager/pkg/requestlog"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

type Service interface {
	List(ctx context.Context, input housesvc.ListInput) (*housesvc.ListResult, error)
	Detail(ctx context.Context, input housesvc.DetailInput) (*housesvc.DetailResult, error)
}

func NewHandler(service *housesvc.Service) *Handler {
	return newHandler(service)
}

func newHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/house/list", middleware.RequirePermission("house.view", "当前账号无权查看房源列表"), h.List)
	rg.POST("/house/detail", middleware.RequirePermission("house.view", "当前账号无权查看房源详情"), h.Detail)
}

func (h *Handler) List(c *gin.Context) {
	var req listRequest
	if err := bindOptionalJSON(c, &req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("请求参数格式不正确")))
		return
	}
	requestlog.AddField(c, "provider_id", req.ProviderID)
	requestlog.AddField(c, "asset_mode", req.AssetMode)

	result, err := h.service.List(c.Request.Context(), housesvc.ListInput{
		ProviderID:    req.ProviderID,
		AssetMode:     req.AssetMode,
		City:          req.City,
		District:      req.District,
		RoomStatus:    req.RoomStatus,
		ListingStatus: req.ListingStatus,
		AuditStatus:   req.AuditStatus,
		Page:          req.Page,
		PageSize:      req.PageSize,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toListResponse(result))
}

func (h *Handler) Detail(c *gin.Context) {
	var req detailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("请求参数格式不正确")))
		return
	}
	requestlog.AddField(c, "listing_id", req.ListingID)

	result, err := h.service.Detail(c.Request.Context(), housesvc.DetailInput{ListingID: req.ListingID})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toDetailResponse(result))
}

var _ handler.RouteRegistrar = (*Handler)(nil)
var _ Service = (*housesvc.Service)(nil)

func bindOptionalJSON(c *gin.Context, dst any) error {
	if c.Request.Body == nil || c.Request.ContentLength == 0 {
		return nil
	}
	if err := c.ShouldBindJSON(dst); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	return nil
}
