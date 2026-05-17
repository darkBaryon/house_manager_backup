package role

import (
	"context"
	"fmt"
	"io"

	"house-manager/internal/handler"
	"house-manager/internal/middleware"
	rolesvc "house-manager/internal/service/admin/role"
	"house-manager/pkg/errcode"
	"house-manager/pkg/response"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

type Service interface {
	List(ctx context.Context, input rolesvc.ListInput) (*rolesvc.ListResult, error)
	Detail(ctx context.Context, input rolesvc.DetailInput) (*rolesvc.DetailResult, error)
	Create(ctx context.Context, input rolesvc.CreateInput) (*rolesvc.CreateResult, error)
	Update(ctx context.Context, input rolesvc.UpdateInput) (*rolesvc.UpdateResult, error)
}

func NewHandler(service *rolesvc.Service) *Handler {
	return newHandler(service)
}

func newHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/role/list", middleware.RequirePermission("role.view", "当前账号无权查看角色列表"), h.List)
	rg.POST("/role/detail", middleware.RequirePermission("role.view", "当前账号无权查看角色详情"), h.Detail)
	rg.POST("/role/create", middleware.RequirePermission("role.edit", "当前账号无权创建角色"), h.Create)
	rg.POST("/role/update", middleware.RequirePermission("role.edit", "当前账号无权编辑角色"), h.Update)
}

func (h *Handler) List(c *gin.Context) {
	var req listRequest
	if err := bindOptionalJSON(c, &req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("请求参数格式不正确")))
		return
	}
	result, err := h.service.List(c.Request.Context(), rolesvc.ListInput{
		Keyword:  req.Keyword,
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
	result, err := h.service.Detail(c.Request.Context(), rolesvc.DetailInput{RoleID: req.RoleID})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toDetailResponse(result))
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
	result, err := h.service.Create(c.Request.Context(), rolesvc.CreateInput{
		OperatorStaffID: principal.PrincipalID,
		RoleName:        req.RoleName,
		RoleCode:        req.RoleCode,
		Description:     req.Description,
		PermissionCodes: req.PermissionCodes,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toCreateResponse(result))
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
	result, err := h.service.Update(c.Request.Context(), rolesvc.UpdateInput{
		OperatorStaffID: principal.PrincipalID,
		RoleID:          req.RoleID,
		RoleName:        req.RoleName,
		Description:     req.Description,
		PermissionCodes: req.PermissionCodes,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toUpdateResponse(result))
}

var _ handler.RouteRegistrar = (*Handler)(nil)
var _ Service = (*rolesvc.Service)(nil)

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
