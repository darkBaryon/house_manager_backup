package hmd

import repohmd "house-manager/internal/repository/hmd"

type Service struct {
	centralizedRepo         *repohmd.CentralizedRepository
	buildingRepo            *repohmd.BuildingRepository
	decentralizedRepo       *repohmd.DecentralizedRepository
	roomTypeCentralizedRepo *repohmd.RoomTypeCentralizedRepository
	roomCentralizedRepo     *repohmd.RoomCentralizedRepository
	roomDecentralizedRepo   *repohmd.RoomDecentralizedRepository
}

func NewService(
	centralizedRepo *repohmd.CentralizedRepository,
	buildingRepo *repohmd.BuildingRepository,
	decentralizedRepo *repohmd.DecentralizedRepository,
	roomTypeCentralizedRepo *repohmd.RoomTypeCentralizedRepository,
	roomCentralizedRepo *repohmd.RoomCentralizedRepository,
	roomDecentralizedRepo *repohmd.RoomDecentralizedRepository,
) *Service {
	return &Service{
		centralizedRepo:         centralizedRepo,
		buildingRepo:            buildingRepo,
		decentralizedRepo:       decentralizedRepo,
		roomTypeCentralizedRepo: roomTypeCentralizedRepo,
		roomCentralizedRepo:     roomCentralizedRepo,
		roomDecentralizedRepo:   roomDecentralizedRepo,
	}
}
