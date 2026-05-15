package publish

import (
	hmddomain "house-manager/internal/domain/hmd"
	"house-manager/internal/domain/listingprojection"
	"house-manager/internal/domain/publishaccess"
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

func NewPublishService(
	hmd *hmddomain.Service,
	listingProjection *listingprojection.Service,
	publishAccess *publishaccess.Service,
) *PublishService {
	publisher := mutationPublisher{listingProjection: listingProjection}
	return &PublishService{
		centralizedProjectService:     newCentralizedProjectService(hmd, publisher, publishAccess),
		buildingService:               newBuildingService(hmd, publisher, publishAccess),
		roomTypeService:               newRoomTypeService(hmd, publisher, publishAccess),
		centralizedRoomService:        newCentralizedRoomService(hmd, publisher, publishAccess),
		decentralizedCommunityService: newDecentralizedCommunityService(hmd, publisher, publishAccess),
		decentralizedRoomService:      newDecentralizedRoomService(hmd, publisher, publishAccess),
	}
}
