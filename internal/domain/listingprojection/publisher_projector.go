package listingprojection

import (
	"context"
	"fmt"

	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PublisherProjector struct {
	hpdListingRepo             hpdListingRepository
	hpdPublisherListingRepo    hpdPublisherListingRepository
	hmdCentralizedRepo         hmdCentralizedRepository
	hmdBuildingRepo            hmdBuildingRepository
	hmdDecentralizedRepo       hmdDecentralizedRepository
	hmdRoomTypeCentralizedRepo hmdRoomTypeCentralizedRepository
	hmdRoomCentralizedRepo     hmdRoomCentralizedRepository
	hmdRoomDecentralizedRepo   hmdRoomDecentralizedRepository
}

func NewPublisherProjector(
	hpdListingRepo hpdListingRepository,
	hpdPublisherListingRepo hpdPublisherListingRepository,
	hmdCentralizedRepo hmdCentralizedRepository,
	hmdBuildingRepo hmdBuildingRepository,
	hmdDecentralizedRepo hmdDecentralizedRepository,
	hmdRoomTypeCentralizedRepo hmdRoomTypeCentralizedRepository,
	hmdRoomCentralizedRepo hmdRoomCentralizedRepository,
	hmdRoomDecentralizedRepo hmdRoomDecentralizedRepository,
) *PublisherProjector {
	return &PublisherProjector{
		hpdListingRepo:             hpdListingRepo,
		hpdPublisherListingRepo:    hpdPublisherListingRepo,
		hmdCentralizedRepo:         hmdCentralizedRepo,
		hmdBuildingRepo:            hmdBuildingRepo,
		hmdDecentralizedRepo:       hmdDecentralizedRepo,
		hmdRoomTypeCentralizedRepo: hmdRoomTypeCentralizedRepo,
		hmdRoomCentralizedRepo:     hmdRoomCentralizedRepo,
		hmdRoomDecentralizedRepo:   hmdRoomDecentralizedRepo,
	}
}

func (p *PublisherProjector) RefreshByListing(ctx context.Context, listing *hpdmodel.HpdListing) error {
	if listing == nil {
		return fmt.Errorf("hpd listing is nil")
	}
	switch listing.SourceType {
	case hpdmodel.HpdSourceTypeCentralizedRoom:
		return p.RefreshCentralizedRoom(ctx, listing.SourceID)
	case hpdmodel.HpdSourceTypeDecentralizedRoom:
		return p.RefreshDecentralizedRoom(ctx, listing.SourceID)
	default:
		return fmt.Errorf("unsupported hpd listing source type: %s", listing.SourceType)
	}
}

func (p *PublisherProjector) RefreshCentralizedRoom(ctx context.Context, roomID bson.ObjectID) error {
	room, err := p.hmdRoomCentralizedRepo.FindByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("find hmd centralized room: %w", err)
	}
	if room == nil {
		return fmt.Errorf("hmd centralized room not found")
	}

	project, err := p.hmdCentralizedRepo.FindByID(ctx, room.ProjectID)
	if err != nil {
		return fmt.Errorf("find hmd centralized project: %w", err)
	}
	if project == nil {
		return fmt.Errorf("hmd centralized project not found")
	}

	building, err := p.hmdBuildingRepo.FindByID(ctx, room.BuildingID)
	if err != nil {
		return fmt.Errorf("find hmd building: %w", err)
	}
	if building == nil {
		return fmt.Errorf("hmd building not found")
	}

	var roomType *hmdmodel.HmdRoomTypeCentralized
	if !room.RoomTypeID.IsZero() {
		roomType, err = p.hmdRoomTypeCentralizedRepo.FindByID(ctx, room.RoomTypeID)
		if err != nil {
			return fmt.Errorf("find hmd centralized room type: %w", err)
		}
	}

	listing, err := p.hpdListingRepo.UpsertBySource(ctx, &hpdmodel.HpdListing{
		SourceType: hpdmodel.HpdSourceTypeCentralizedRoom,
		SourceID:   room.ID,
		AssetMode:  hpdmodel.HpdAssetModeCentralized,
	})
	if err != nil {
		return fmt.Errorf("upsert hpd listing: %w", err)
	}
	if listing == nil {
		return fmt.Errorf("upsert hpd listing returned nil")
	}

	publisherListing := mapCentralizedPublisherListing(listing, room, project, building, roomType)
	if _, err := p.hpdPublisherListingRepo.UpsertByListingID(ctx, publisherListing); err != nil {
		return fmt.Errorf("upsert hpd publisher listing: %w", err)
	}
	return nil
}

func (p *PublisherProjector) RefreshDecentralizedRoom(ctx context.Context, roomID bson.ObjectID) error {
	room, err := p.hmdRoomDecentralizedRepo.FindByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("find hmd decentralized room: %w", err)
	}
	if room == nil {
		return fmt.Errorf("hmd decentralized room not found")
	}

	community, err := p.hmdDecentralizedRepo.FindByID(ctx, room.DecentralizedID)
	if err != nil {
		return fmt.Errorf("find hmd decentralized community: %w", err)
	}
	if community == nil {
		return fmt.Errorf("hmd decentralized community not found")
	}

	listing, err := p.hpdListingRepo.UpsertBySource(ctx, &hpdmodel.HpdListing{
		SourceType: hpdmodel.HpdSourceTypeDecentralizedRoom,
		SourceID:   room.ID,
		AssetMode:  hpdmodel.HpdAssetModeDecentralized,
	})
	if err != nil {
		return fmt.Errorf("upsert hpd listing: %w", err)
	}
	if listing == nil {
		return fmt.Errorf("upsert hpd listing returned nil")
	}

	publisherListing := mapDecentralizedPublisherListing(listing, room, community)
	if _, err := p.hpdPublisherListingRepo.UpsertByListingID(ctx, publisherListing); err != nil {
		return fmt.Errorf("upsert hpd publisher listing: %w", err)
	}
	return nil
}
