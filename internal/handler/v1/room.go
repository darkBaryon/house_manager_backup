package v1

import (
	"context"

	"house-manager/internal/handler"
	"house-manager/internal/model"
	"house-manager/pkg/errcode"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

// RoomServicer 房间业务接口（handler 消费 service）
type RoomServicer interface {
	GetDetail(ctx context.Context, id string) (*model.Room, error)
	GetList(ctx context.Context, req model.PageReq) ([]model.Room, int64, error)
}

// RoomHandler 房间处理器
type RoomHandler struct {
	Svc RoomServicer
}

// NewRoomHandler 创建房间处理器
func NewRoomHandler(svc RoomServicer) *RoomHandler {
	return &RoomHandler{Svc: svc}
}

func (h *RoomHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/room/detail", h.Detail)
	rg.POST("/room/list", h.List)
}

func (h *RoomHandler) Detail(c *gin.Context) {
	var req struct {
		Id string `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam)
		return
	}

	room, err := h.Svc.GetDetail(c.Request.Context(), req.Id)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, room)
}

func (h *RoomHandler) List(c *gin.Context) {
	var req model.PageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, errcode.InvalidParam)
		return
	}

	rooms, total, err := h.Svc.GetList(c.Request.Context(), req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.SuccessPage(c, rooms, len(rooms), total)
}

var _ handler.RouteRegistrar = (*RoomHandler)(nil)
