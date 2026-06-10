package provider

import (
	"context"
	"fmt"
	"io"

	"house-manager/internal/handler"
	"house-manager/internal/middleware"
	providersvc "house-manager/internal/service/admin/provider"
	"house-manager/pkg/applog"
	"house-manager/pkg/errcode"
	"house-manager/pkg/requestlog"
	"house-manager/pkg/response"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

type Service interface {
	Create(ctx context.Context, input providersvc.CreateInput) (*providersvc.CreateResult, error)
	List(ctx context.Context, input providersvc.ListInput) (*providersvc.ListResult, error)
	Detail(ctx context.Context, input providersvc.DetailInput) (*providersvc.DetailResult, error)
	Update(ctx context.Context, input providersvc.UpdateInput) (*providersvc.UpdateResult, error)
	Disable(ctx context.Context, input providersvc.DisableInput) (*providersvc.DisableResult, error)
}

func NewHandler(service *providersvc.Service) *Handler {
	return newHandler(service)
}

func newHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/provider/list", middleware.RequirePermission("provider.view", "当前账号无权查看发房方列表"), h.List)
	rg.POST("/provider/detail", middleware.RequirePermission("provider.view", "当前账号无权查看发房方详情"), h.Detail)
	rg.POST("/provider/create", middleware.RequirePermission("provider.edit", "当前账号无权创建发房方"), h.Create)
	rg.POST("/provider/update", middleware.RequirePermission("provider.edit", "当前账号无权编辑发房方"), h.Update)
	rg.POST("/provider/disable", middleware.RequirePermission("provider.edit", "当前账号无权禁用发房方"), h.Disable)
}

func (h *Handler) Create(c *gin.Context) {
	principal, ok := session.PrincipalFromContext(c.Request.Context())
	if !ok {
		response.Err(c, errcode.Unauthorized.WithError(fmt.Errorf("未登录或登录已过期，请重新登录")))
		return
	}

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("请求参数格式不正确")))
		return
	}
	requestlog.AddField(c, "phone", applog.MaskPhone(req.Phone))

	result, err := h.service.Create(c.Request.Context(), providersvc.CreateInput{
		OperatorStaffID: principal.PrincipalID,
		Phone:           req.Phone,
		Password:        req.Password,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toCreateResponse(result))
}

func (h *Handler) List(c *gin.Context) {
	var req listRequest
	if err := bindOptionalJSON(c, &req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("请求参数格式不正确")))
		return
	}

	result, err := h.service.List(c.Request.Context(), providersvc.ListInput{
		Phone:    req.Phone,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
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

	result, err := h.service.Detail(c.Request.Context(), providersvc.DetailInput{ProviderID: req.ProviderID})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toDetailResponse(result))
}

func (h *Handler) Update(c *gin.Context) {
	principal, ok := session.PrincipalFromContext(c.Request.Context())
	if !ok {
		response.Err(c, errcode.Unauthorized.WithError(fmt.Errorf("未登录或登录已过期，请重新登录")))
		return
	}

	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("请求参数格式不正确")))
		return
	}
	requestlog.AddField(c, "provider_id", req.ProviderID)
	requestlog.AddField(c, "phone", applog.MaskPhone(req.Phone))

	result, err := h.service.Update(c.Request.Context(), providersvc.UpdateInput{
		OperatorStaffID: principal.PrincipalID,
		ProviderID:      req.ProviderID,
		Phone:           req.Phone,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toUpdateResponse(result))
}

func (h *Handler) Disable(c *gin.Context) {
	principal, ok := session.PrincipalFromContext(c.Request.Context())
	if !ok {
		response.Err(c, errcode.Unauthorized.WithError(fmt.Errorf("未登录或登录已过期，请重新登录")))
		return
	}

	var req disableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("请求参数格式不正确")))
		return
	}
	requestlog.AddField(c, "provider_id", req.ProviderID)

	result, err := h.service.Disable(c.Request.Context(), providersvc.DisableInput{
		OperatorStaffID: principal.PrincipalID,
		ProviderID:      req.ProviderID,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toDisableResponse(result))
}

var _ handler.RouteRegistrar = (*Handler)(nil)
var _ Service = (*providersvc.Service)(nil)

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
