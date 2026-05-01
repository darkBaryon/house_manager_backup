package v2

import (
	"strconv"

	"house-manager/internal/handler"
	v1handler "house-manager/internal/handler/v1"
	"house-manager/internal/model"
	"house-manager/pkg/errcode"
	"house-manager/pkg/response"

	"github.com/gin-gonic/gin"
)

// RoomHandler v2 RESTful 房间接口（GET /rooms/:id, GET /rooms）
type RoomHandler struct {
	Svc v1handler.RoomServicer
}

// NewRoomHandler 创建 v2 房间处理器
func NewRoomHandler(svc v1handler.RoomServicer) *RoomHandler {
	return &RoomHandler{Svc: svc}
}

func (h *RoomHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rooms := rg.Group("/rooms")
	rooms.GET("/:id", h.GetDetail)
	rooms.GET("", h.GetList)
}

func (h *RoomHandler) GetDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Err(c, errcode.InvalidParam)
		return
	}

	room, err := h.Svc.GetDetail(c.Request.Context(), id)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, room)
}

func (h *RoomHandler) GetList(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	req := model.PageReq{
		Offset: offset,
		Limit:  limit,
	}

	rooms, total, err := h.Svc.GetList(c.Request.Context(), req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.SuccessPage(c, rooms, len(rooms), total)
}

var _ handler.RouteRegistrar = (*RoomHandler)(nil)
