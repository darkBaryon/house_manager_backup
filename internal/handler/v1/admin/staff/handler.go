package staff

import (
	"context"
	"fmt"
	"io"

	"house-manager/internal/handler"
	"house-manager/internal/middleware"
	staffsvc "house-manager/internal/service/admin/staff"
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
	Create(ctx context.Context, input staffsvc.CreateInput) (*staffsvc.CreateResult, error)
	List(ctx context.Context, input staffsvc.ListInput) (*staffsvc.ListResult, error)
	Detail(ctx context.Context, input staffsvc.DetailInput) (*staffsvc.DetailResult, error)
	Update(ctx context.Context, input staffsvc.UpdateInput) (*staffsvc.UpdateResult, error)
	Disable(ctx context.Context, input staffsvc.DisableInput) (*staffsvc.DisableResult, error)
}

func NewHandler(service *staffsvc.Service) *Handler {
	return newHandler(service)
}

func newHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/staff/create", middleware.RequirePermission("staff.edit", "当前账号无权创建员工"), h.Create)
	rg.POST("/staff/list", middleware.RequirePermission("staff.view", "当前账号无权查看员工列表"), h.List)
	rg.POST("/staff/detail", middleware.RequirePermission("staff.view", "当前账号无权查看员工详情"), h.Detail)
	rg.POST("/staff/update", middleware.RequirePermission("staff.edit", "当前账号无权编辑员工"), h.Update)
	rg.POST("/staff/disable", middleware.RequirePermission("staff.edit", "当前账号无权禁用员工"), h.Disable)
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

	result, err := h.service.Create(c.Request.Context(), staffsvc.CreateInput{
		OperatorStaffID: principal.PrincipalID,
		Name:            req.Name,
		Phone:           req.Phone,
		Password:        req.Password,
		Email:           req.Email,
		Department:      req.Department,
		JobTitle:        req.JobTitle,
		ContactQRCode:   req.ContactQRCode,
		RoleIDs:         req.RoleIDs,
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

	result, err := h.service.List(c.Request.Context(), staffsvc.ListInput{
		Keyword:  req.Keyword,
		Phone:    req.Phone,
		RoleID:   req.RoleID,
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

	result, err := h.service.Detail(c.Request.Context(), staffsvc.DetailInput{
		StaffID: req.StaffID,
	})
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

	result, err := h.service.Update(c.Request.Context(), staffsvc.UpdateInput{
		OperatorStaffID: principal.PrincipalID,
		StaffID:         req.StaffID,
		Name:            req.Name,
		Email:           req.Email,
		Department:      req.Department,
		JobTitle:        req.JobTitle,
		ContactQRCode:   req.ContactQRCode,
		Status:          req.Status,
		RoleIDs:         req.RoleIDs,
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

	result, err := h.service.Disable(c.Request.Context(), staffsvc.DisableInput{
		OperatorStaffID: principal.PrincipalID,
		StaffID:         req.StaffID,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toDisableResponse(result))
}

var _ handler.RouteRegistrar = (*Handler)(nil)
var _ Service = (*staffsvc.Service)(nil)

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
