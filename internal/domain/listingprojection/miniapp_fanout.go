package listingprojection

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	hmdmodel "house-manager/internal/model/hmd"
)

func (p *MiniappProjector) RefreshCentralizedProject(ctx context.Context, projectID bson.ObjectID) error {
	rooms, err := p.hmdRoomCentralizedRepo.ListByProjectID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("list centralized rooms by project: %w", err)
	}
	return p.refreshCentralizedRooms(ctx, rooms)
}

func (p *MiniappProjector) RefreshBuilding(ctx context.Context, buildingID bson.ObjectID) error {
	rooms, err := p.hmdRoomCentralizedRepo.ListByBuildingID(ctx, buildingID)
	if err != nil {
		return fmt.Errorf("list centralized rooms by building: %w", err)
	}
	return p.refreshCentralizedRooms(ctx, rooms)
}

func (p *MiniappProjector) RefreshRoomTypeCentralized(ctx context.Context, roomTypeID bson.ObjectID) error {
	rooms, err := p.hmdRoomCentralizedRepo.ListByRoomTypeID(ctx, roomTypeID)
	if err != nil {
		return fmt.Errorf("list centralized rooms by room type: %w", err)
	}
	return p.refreshCentralizedRooms(ctx, rooms)
}

func (p *MiniappProjector) RefreshDecentralizedCommunity(ctx context.Context, decentralizedID bson.ObjectID) error {
	rooms, err := p.hmdRoomDecentralizedRepo.ListByDecentralizedID(ctx, decentralizedID)
	if err != nil {
		return fmt.Errorf("list decentralized rooms by community: %w", err)
	}
	for _, room := range rooms {
		if err := p.RefreshDecentralizedRoom(ctx, room.ID); err != nil {
			return fmt.Errorf("refresh decentralized room %s: %w", room.ID.Hex(), err)
		}
	}
	return nil
}

func (p *MiniappProjector) refreshCentralizedRooms(ctx context.Context, rooms []hmdmodel.HmdRoomCentralized) error {
	for _, room := range rooms {
		if err := p.RefreshCentralizedRoom(ctx, room.ID); err != nil {
			return fmt.Errorf("refresh centralized room %s: %w", room.ID.Hex(), err)
		}
	}
	return nil
}
