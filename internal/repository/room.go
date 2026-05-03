package repository

import (
	"house-manager/internal/model"
	dbmongo "house-manager/pkg/database/mongo"

)

// RoomRepository 房间数据仓库（嵌入泛型 Repository 获得通用查询能力）
type RoomRepository struct {
	*Repository[model.Room]
}

// NewRoomRepository 创建房间仓库
func NewRoomRepository(client *dbmongo.Client) *RoomRepository {
	return &RoomRepository{
		Repository: NewRepository[model.Room](client.Collection("room")),
	}
}
