package auth

import (
	"context"
	"log/slog"

	"house-manager/internal/handler"
	authsvc "house-manager/internal/service/miniapp/auth"
	"house-manager/pkg/errcode"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service authService
}

type SessionHandler struct{}

type authService interface {
	WechatLogin(ctx context.Context, code, loginIP string) (string, error)
	WechatRegister(ctx context.Context, code, phoneCode, loginIP string) (string, error)
}

type wechatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

type wechatRegisterRequest struct {
	Code      string `json:"code" binding:"required"`
	PhoneCode string `json:"phone_code" binding:"required"`
}

func NewAuthHandler(service *authsvc.Service) *AuthHandler {
	return newAuthHandler(service)
}

func newAuthHandler(service authService) *AuthHandler {
	return &AuthHandler{service: service}
}

func NewSessionHandler() *SessionHandler {
	return &SessionHandler{}
}

func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/auth/wechat_login", h.WechatLogin)
	rg.POST("/auth/wechat_register", h.WechatRegister)
}

func (h *SessionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/auth/session", h.Session)
}

func (h *AuthHandler) WechatLogin(c *gin.Context) {
	var req wechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(err))
		return
	}

	token, err := h.service.WechatLogin(c.Request.Context(), req.Code, c.ClientIP())
	if err != nil {
		slog.Error("wechat login failed", "error", err)
		response.Err(c, err)
		return
	}
	response.Success(c, gin.H{"token": token})
}

func (h *AuthHandler) WechatRegister(c *gin.Context) {
	var req wechatRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(err))
		return
	}

	token, err := h.service.WechatRegister(c.Request.Context(), req.Code, req.PhoneCode, c.ClientIP())
	if err != nil {
		slog.Error("wechat register failed", "error", err)
		response.Err(c, err)
		return
	}
	response.Success(c, gin.H{"token": token})
}

func (h *SessionHandler) Session(c *gin.Context) {
	userId, ok := c.Get("userId")
	if !ok {
		response.Err(c, errcode.Unauthorized)
		return
	}
	response.Success(c, gin.H{"userId": userId})
}

var _ handler.RouteRegistrar = (*AuthHandler)(nil)
var _ handler.RouteRegistrar = (*SessionHandler)(nil)
var _ authService = (*authsvc.Service)(nil)
