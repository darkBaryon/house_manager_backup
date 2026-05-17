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
	ListRoots(ctx context.Context, input housesvc.RootListInput) (*housesvc.RootListResult, error)
	ListBuildings(ctx context.Context, input housesvc.BuildingListInput) (*housesvc.BuildingListResult, error)
	ListRooms(ctx context.Context, input housesvc.ListInput) (*housesvc.ListResult, error)
	DetailRoom(ctx context.Context, input housesvc.DetailInput) (*housesvc.DetailResult, error)
}

func NewHandler(service *housesvc.Service) *Handler {
	return newHandler(service)
}

func newHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/house_root/list", middleware.RequirePermission("house.view", "当前账号无权查看项目/小区列表"), h.ListRoots)
	rg.POST("/house_building/list", middleware.RequirePermission("house.view", "当前账号无权查看楼栋列表"), h.ListBuildings)
	rg.POST("/house_room/list", middleware.RequirePermission("house.view", "当前账号无权查看房间列表"), h.ListRooms)
	rg.POST("/house_room/detail", middleware.RequirePermission("house.view", "当前账号无权查看房间详情"), h.DetailRoom)
}

func (h *Handler) ListRoots(c *gin.Context) {
	var req rootListRequest
	if err := bindOptionalJSON(c, &req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("请求参数格式不正确")))
		return
	}
	requestlog.AddField(c, "provider_id", req.ProviderID)
	requestlog.AddField(c, "asset_mode", req.AssetMode)

	result, err := h.service.ListRoots(c.Request.Context(), housesvc.RootListInput{
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
	response.Success(c, toRootListResponse(result))
}

func (h *Handler) ListBuildings(c *gin.Context) {
	var req buildingListRequest
	if err := bindOptionalJSON(c, &req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("请求参数格式不正确")))
		return
	}
	requestlog.AddField(c, "root_id", req.RootID)

	result, err := h.service.ListBuildings(c.Request.Context(), housesvc.BuildingListInput{
		RootID:        req.RootID,
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
	response.Success(c, toBuildingListResponse(result))
}

func (h *Handler) ListRooms(c *gin.Context) {
	var req roomListRequest
	if err := bindOptionalJSON(c, &req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("请求参数格式不正确")))
		return
	}
	requestlog.AddField(c, "root_id", req.RootID)
	requestlog.AddField(c, "building_id", req.BuildingID)

	result, err := h.service.ListRooms(c.Request.Context(), housesvc.ListInput{
		RootID:        req.RootID,
		BuildingID:    req.BuildingID,
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
	response.Success(c, toRoomListResponse(result))
}

func (h *Handler) DetailRoom(c *gin.Context) {
	var req detailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("请求参数格式不正确")))
		return
	}
	requestlog.AddField(c, "listing_id", req.ListingID)

	result, err := h.service.DetailRoom(c.Request.Context(), housesvc.DetailInput{ListingID: req.ListingID})
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
