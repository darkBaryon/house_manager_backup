package wire

import (
	"house-manager/internal/config"
	v1handler "house-manager/internal/handler/v1"
	v2handler "house-manager/internal/handler/v2"
	"house-manager/internal/repository"
	"house-manager/internal/service"

	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type dbName string

func newDbName(cfg *config.Config) dbName {
	return dbName(cfg.MongoDB.Database)
}

func newRoomRepository(client *mongo.Client, name dbName) *repository.RoomRepository {
	return repository.NewRoomRepository(client, string(name))
}

var RoomSet = wire.NewSet(
	newDbName,
	newRoomRepository,
	newCache,
	service.NewRoomService,
	v1handler.NewRoomHandler,
	v2handler.NewRoomHandler,
	wire.Bind(new(service.RoomFinder), new(*repository.RoomRepository)),
	wire.Bind(new(v1handler.RoomServicer), new(*service.RoomService)),
)
