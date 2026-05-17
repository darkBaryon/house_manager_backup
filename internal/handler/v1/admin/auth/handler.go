package auth

import (
	"context"
	"fmt"

	"house-manager/internal/handler"
	"house-manager/internal/middleware"
	authsvc "house-manager/internal/service/admin/auth"
	"house-manager/pkg/errcode"
	"house-manager/pkg/requestlog"
	"house-manager/pkg/response"
	"house-manager/pkg/session"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

type PublicHandler struct {
	service Service
}

type Service interface {
	Login(ctx context.Context, input authsvc.LoginInput) (*authsvc.LoginResult, error)
	Session(ctx context.Context, principal session.Principal) (*authsvc.AuthSession, error)
	Logout(ctx context.Context, token string) error
}

func NewHandler(service *authsvc.Service) *Handler {
	return newHandler(service)
}

func NewPublicHandler(service *authsvc.Service) *PublicHandler {
	return &PublicHandler{service: service}
}

func newHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *PublicHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/admin_auth/login", h.Login)
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/admin_auth/session", h.Session)
	rg.POST("/admin_auth/logout", h.Logout)
}

func (h *PublicHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(fmt.Errorf("请求参数格式不正确")))
		return
	}
	requestlog.AddField(c, "phone", maskPhone(req.Phone))
	result, err := h.service.Login(c.Request.Context(), authsvc.LoginInput{
		Phone:     req.Phone,
		Password:  req.Password,
		LoginIP:   c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		requestlog.AddField(c, "action", "admin login failed")
		response.Err(c, err)
		return
	}
	response.Success(c, toLoginResponse(result))
}

func (h *Handler) Session(c *gin.Context) {
	principal, ok := principalFromContext(c)
	if !ok {
		response.Err(c, errcode.Unauthorized.WithError(fmt.Errorf("未登录或登录已过期，请重新登录")))
		return
	}
	result, err := h.service.Session(c.Request.Context(), principal)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toSessionResponse(result))
}

func (h *Handler) Logout(c *gin.Context) {
	value, ok := c.Get(middleware.ContextToken)
	token, ok := value.(string)
	if !ok || token == "" {
		response.Err(c, errcode.Unauthorized.WithError(fmt.Errorf("未登录或登录已过期，请重新登录")))
		return
	}
	if err := h.service.Logout(c.Request.Context(), token); err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, logoutResponse{Success: true})
}

func principalFromContext(c *gin.Context) (session.Principal, bool) {
	value, ok := c.Get(middleware.ContextPrincipal)
	if !ok {
		return session.Principal{}, false
	}
	principal, ok := value.(session.Principal)
	return principal, ok
}

var _ handler.RouteRegistrar = (*Handler)(nil)
var _ handler.RouteRegistrar = (*PublicHandler)(nil)
var _ Service = (*authsvc.Service)(nil)

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
