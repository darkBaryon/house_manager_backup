package publish

import (
	"context"

	hmdsvc "house-manager/internal/service/hmd"
	hpdsvc "house-manager/internal/service/hpd"
)

// PublishService 是发房域的业务入口，handler 只依赖这一层。
type PublishService struct {
	centralizedProjects      centralizedProjectService
	buildings                buildingService
	roomTypes                roomTypeService
	centralizedRooms         centralizedRoomService
	decentralizedCommunities decentralizedCommunityService
	decentralizedRooms       decentralizedRoomService
	hpd                      hpdApplier
}

func NewPublishService(hmd *hmdsvc.Service, hpd *hpdsvc.Service) *PublishService {
	return &PublishService{
		centralizedProjects:      hmd,
		buildings:                hmd,
		roomTypes:                hmd,
		centralizedRooms:         hmd,
		decentralizedCommunities: hmd,
		decentralizedRooms:       hmd,
		hpd:                      hpd,
	}
}

func resolveHmdMutation[T any](ctx context.Context, hpd hpdApplier, result *hmdsvc.HmdMutationResult[T], err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	if err := hpd.Apply(ctx, result.Changes); err != nil {
		return nil, err
	}
	return result.Entity, nil
}
