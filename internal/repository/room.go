package repository

import (
	"house-manager/internal/model"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// RoomRepository 房间数据仓库（嵌入泛型 Repository 获得通用查询能力）
type RoomRepository struct {
	*Repository[model.Room]
}

// NewRoomRepository 创建房间仓库
func NewRoomRepository(client *mongo.Client, dbName string) *RoomRepository {
	return &RoomRepository{
		Repository: NewRepository[model.Room](
			client.Database(dbName).Collection("room"),
		),
	}
}
