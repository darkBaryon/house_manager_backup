package listingprojection

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"house-manager/internal/domain/hmd"
	hmdmodel "house-manager/internal/model/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	"log/slog"
)

type MiniappProjector struct {
	hpdListingRepo             hpdListingRepository
	hpdMiniappListingRepo      hpdMiniappListingRepository
	hmdCentralizedRepo         hmdCentralizedRepository
	hmdBuildingRepo            hmdBuildingRepository
	hmdDecentralizedRepo       hmdDecentralizedRepository
	hmdRoomTypeCentralizedRepo hmdRoomTypeCentralizedRepository
	hmdRoomCentralizedRepo     hmdRoomCentralizedRepository
	hmdRoomDecentralizedRepo   hmdRoomDecentralizedRepository
}

func NewMiniappProjector(
	hpdListingRepo hpdListingRepository,
	hpdMiniappListingRepo hpdMiniappListingRepository,
	hmdCentralizedRepo hmdCentralizedRepository,
	hmdBuildingRepo hmdBuildingRepository,
	hmdDecentralizedRepo hmdDecentralizedRepository,
	hmdRoomTypeCentralizedRepo hmdRoomTypeCentralizedRepository,
	hmdRoomCentralizedRepo hmdRoomCentralizedRepository,
	hmdRoomDecentralizedRepo hmdRoomDecentralizedRepository,
) *MiniappProjector {
	return &MiniappProjector{
		hpdListingRepo:             hpdListingRepo,
		hpdMiniappListingRepo:      hpdMiniappListingRepo,
		hmdCentralizedRepo:         hmdCentralizedRepo,
		hmdBuildingRepo:            hmdBuildingRepo,
		hmdDecentralizedRepo:       hmdDecentralizedRepo,
		hmdRoomTypeCentralizedRepo: hmdRoomTypeCentralizedRepo,
		hmdRoomCentralizedRepo:     hmdRoomCentralizedRepo,
		hmdRoomDecentralizedRepo:   hmdRoomDecentralizedRepo,
	}
}

func (p *MiniappProjector) RefreshByListing(ctx context.Context, listing *hpdmodel.HpdListing) error {
	if listing == nil {
		return fmt.Errorf("hpd listing is nil")
	}
	slog.InfoContext(ctx, "listingprojection.miniapp.refresh_by_listing.start", "listing_id", listing.ID.Hex(), "source_type", listing.SourceType, "source_id", listing.SourceID.Hex())
	switch listing.SourceType {
	case hpdmodel.HpdSourceTypeCentralizedRoom:
		return p.RefreshCentralizedRoom(ctx, listing.SourceID)
	case hpdmodel.HpdSourceTypeDecentralizedRoom:
		return p.RefreshDecentralizedRoom(ctx, listing.SourceID)
	default:
		return fmt.Errorf("unsupported hpd listing source type: %s", listing.SourceType)
	}
}

func (p *MiniappProjector) RefreshCentralizedRoom(ctx context.Context, roomID bson.ObjectID) error {
	slog.InfoContext(ctx, "listingprojection.miniapp.refresh_centralized_room.start", "room_id", roomID.Hex())
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

	miniappListing := mapCentralizedMiniappListing(listing, room, project, building, roomType)
	if _, err := p.hpdMiniappListingRepo.UpsertByListingID(ctx, miniappListing); err != nil {
		return fmt.Errorf("upsert hpd miniapp listing: %w", err)
	}
	slog.InfoContext(ctx, "listingprojection.miniapp.refresh_centralized_room.success", "room_id", roomID.Hex(), "listing_id", listing.ID.Hex())
	return nil
}

func (p *MiniappProjector) RefreshDecentralizedRoom(ctx context.Context, roomID bson.ObjectID) error {
	slog.InfoContext(ctx, "listingprojection.miniapp.refresh_decentralized_room.start", "room_id", roomID.Hex())
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

	miniappListing := mapDecentralizedMiniappListing(listing, room, community)
	if _, err := p.hpdMiniappListingRepo.UpsertByListingID(ctx, miniappListing); err != nil {
		return fmt.Errorf("upsert hpd miniapp listing: %w", err)
	}
	slog.InfoContext(ctx, "listingprojection.miniapp.refresh_decentralized_room.success", "room_id", roomID.Hex(), "listing_id", listing.ID.Hex())
	return nil
}

// Refresh 实现 projector 窄接口：声明本端 scope→实体方法的绑定，
// 路由与未知 scope 兜底由包内共享的 dispatchRefresh 完成。
func (p *MiniappProjector) Refresh(ctx context.Context, change hmd.HmdChange) error {
	return dispatchRefresh(ctx, "listingprojection.miniapp.refresh.unhandled_scope", change, p.bindings())
}

// bindings 声明本端 scope→实体方法绑定；完整性由 TestProjectorBindingsComplete
// 反射护栏保证（keyed 字面量漏绑定可编译、运行时才暴露，故须测试兜底）。
func (p *MiniappProjector) bindings() refreshFuncs {
	return refreshFuncs{
		CentralizedProject:     p.RefreshCentralizedProject,
		Building:               p.RefreshBuilding,
		RoomTypeCentralized:    p.RefreshRoomTypeCentralized,
		CentralizedRoom:        p.RefreshCentralizedRoom,
		DecentralizedCommunity: p.RefreshDecentralizedCommunity,
		DecentralizedRoom:      p.RefreshDecentralizedRoom,
	}
}
