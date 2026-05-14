package publish

import (
	hmddomain "house-manager/internal/domain/hmd"
	hpddomain "house-manager/internal/domain/hpd"
)

// PublishService 是发房域的业务入口，handler 只依赖这一层。
type PublishService struct {
	*centralizedProjectService
	*buildingService
	*roomTypeService
	*centralizedRoomService
	*decentralizedCommunityService
	*decentralizedRoomService
}

func NewPublishService(hmd *hmddomain.Service, hpd *hpddomain.Service) *PublishService {
	publisher := mutationPublisher{hpd: hpd}
	return &PublishService{
		centralizedProjectService:     newCentralizedProjectService(hmd, publisher),
		buildingService:               newBuildingService(hmd, publisher),
		roomTypeService:               newRoomTypeService(hmd, publisher),
		centralizedRoomService:        newCentralizedRoomService(hmd, publisher),
		decentralizedCommunityService: newDecentralizedCommunityService(hmd, publisher),
		decentralizedRoomService:      newDecentralizedRoomService(hmd, publisher),
	}
}
