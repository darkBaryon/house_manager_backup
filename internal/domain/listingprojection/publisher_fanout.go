package listingprojection

import (
	"context"
	"fmt"

	hmdmodel "house-manager/internal/model/hmd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (p *PublisherProjector) RefreshCentralizedProject(ctx context.Context, projectID bson.ObjectID) error {
	logProjectionInfo(ctx, "listingprojection.publisher.refresh_project.start", "project_id", projectID.Hex())
	rooms, err := p.hmdRoomCentralizedRepo.ListByProjectID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("list centralized rooms by project: %w", err)
	}
	logProjectionInfo(ctx, "listingprojection.publisher.refresh_project.fanout", "project_id", projectID.Hex(), "room_count", len(rooms))
	return p.refreshCentralizedRooms(ctx, rooms)
}

func (p *PublisherProjector) RefreshBuilding(ctx context.Context, buildingID bson.ObjectID) error {
	logProjectionInfo(ctx, "listingprojection.publisher.refresh_building.start", "building_id", buildingID.Hex())
	rooms, err := p.hmdRoomCentralizedRepo.ListByBuildingID(ctx, buildingID)
	if err != nil {
		return fmt.Errorf("list centralized rooms by building: %w", err)
	}
	logProjectionInfo(ctx, "listingprojection.publisher.refresh_building.fanout", "building_id", buildingID.Hex(), "room_count", len(rooms))
	return p.refreshCentralizedRooms(ctx, rooms)
}

func (p *PublisherProjector) RefreshRoomTypeCentralized(ctx context.Context, roomTypeID bson.ObjectID) error {
	logProjectionInfo(ctx, "listingprojection.publisher.refresh_room_type.start", "room_type_id", roomTypeID.Hex())
	rooms, err := p.hmdRoomCentralizedRepo.ListByRoomTypeID(ctx, roomTypeID)
	if err != nil {
		return fmt.Errorf("list centralized rooms by room type: %w", err)
	}
	logProjectionInfo(ctx, "listingprojection.publisher.refresh_room_type.fanout", "room_type_id", roomTypeID.Hex(), "room_count", len(rooms))
	return p.refreshCentralizedRooms(ctx, rooms)
}

func (p *PublisherProjector) RefreshDecentralizedCommunity(ctx context.Context, decentralizedID bson.ObjectID) error {
	logProjectionInfo(ctx, "listingprojection.publisher.refresh_community.start", "community_id", decentralizedID.Hex())
	rooms, err := p.hmdRoomDecentralizedRepo.ListByDecentralizedID(ctx, decentralizedID)
	if err != nil {
		return fmt.Errorf("list decentralized rooms by community: %w", err)
	}
	logProjectionInfo(ctx, "listingprojection.publisher.refresh_community.fanout", "community_id", decentralizedID.Hex(), "room_count", len(rooms))
	for _, room := range rooms {
		if err := p.RefreshDecentralizedRoom(ctx, room.ID); err != nil {
			return fmt.Errorf("refresh decentralized room %s: %w", room.ID.Hex(), err)
		}
	}
	return nil
}

func (p *PublisherProjector) refreshCentralizedRooms(ctx context.Context, rooms []hmdmodel.HmdRoomCentralized) error {
	for _, room := range rooms {
		if err := p.RefreshCentralizedRoom(ctx, room.ID); err != nil {
			return fmt.Errorf("refresh centralized room %s: %w", room.ID.Hex(), err)
		}
	}
	return nil
}
