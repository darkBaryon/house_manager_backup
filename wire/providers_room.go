package wire

import (
	v1handler "house-manager/internal/handler/v1"
	"house-manager/internal/repository"
	"house-manager/internal/service"
	dbmongo "house-manager/pkg/database/mongo"

	"github.com/google/wire"
)

func newRoomRepository(client *dbmongo.Client) *repository.RoomRepository {
	return repository.NewRoomRepository(client)
}

var RoomSet = wire.NewSet(
	newRoomRepository,
	service.NewRoomService,
	v1handler.NewRoomHandler,
	wire.Bind(new(service.RoomFinder), new(*repository.RoomRepository)),
	wire.Bind(new(v1handler.RoomServicer), new(*service.RoomService)),
)
