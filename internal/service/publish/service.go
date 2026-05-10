package publish

import (
	"context"

	hmddomain "house-manager/internal/domain/hmd"
	hpddomain "house-manager/internal/domain/hpd"
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

func NewPublishService(hmd *hmddomain.Service, hpd *hpddomain.Service) *PublishService {
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

func resolveHmdMutation[T any](ctx context.Context, hpd hpdApplier, result *hmddomain.HmdMutationResult[T], err error) (*T, error) {
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
