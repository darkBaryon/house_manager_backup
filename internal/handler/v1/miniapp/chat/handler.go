package chat

import (
	"context"

	"house-manager/internal/handler"
	minihandler "house-manager/internal/handler/v1/miniapp/common"
	chatsvc "house-manager/internal/service/chat"
	"house-manager/pkg/errcode"
	"house-manager/pkg/requestlog"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service chatService
}

type chatService interface {
	Send(ctx context.Context, input chatsvc.SendInput) (*chatsvc.SendResult, error)
}

func NewHandler(service *chatsvc.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/chat/send", h.Send)
}

func (h *Handler) Send(c *gin.Context) {
	var req sendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam.WithError(err))
		return
	}
	userID, ok := minihandler.CurrentUserID(c)
	if !ok {
		return
	}

	result, err := h.service.Send(c.Request.Context(), req.toServiceInput(requestlog.RequestID(c), userID.Hex()))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, toSendResponse(result))
}

var _ handler.RouteRegistrar = (*Handler)(nil)
var _ chatService = (*chatsvc.Service)(nil)
