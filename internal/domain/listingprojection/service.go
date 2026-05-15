package listingprojection

import (
	"context"

	"house-manager/internal/domain/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	repohpd "house-manager/internal/repository/hpd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Service consumes HMD mutations and refreshes HPD listing read models.
// The current concrete projection target is the miniapp read model; publish ownership
// relations live in package publishaccess instead of this projection service.
type Service struct {
	listingRepo      hpdListingRepository
	miniappProjector miniappProjector
}

func NewService(
	hpdListingRepo *repohpd.ListingRepository,
	miniappProjector *MiniappProjector,
) *Service {
	return &Service{
		listingRepo:      hpdListingRepo,
		miniappProjector: miniappProjector,
	}
}

type miniappProjector interface {
	RefreshByListing(ctx context.Context, listing *hpdmodel.HpdListing) error
	RefreshCentralizedRoom(ctx context.Context, roomID bson.ObjectID) error
	RefreshDecentralizedRoom(ctx context.Context, roomID bson.ObjectID) error
	RefreshCentralizedProject(ctx context.Context, projectID bson.ObjectID) error
	RefreshBuilding(ctx context.Context, buildingID bson.ObjectID) error
	RefreshRoomTypeCentralized(ctx context.Context, roomTypeID bson.ObjectID) error
	RefreshDecentralizedCommunity(ctx context.Context, decentralizedID bson.ObjectID) error
}

func (s *Service) Apply(ctx context.Context, changes []hmd.HmdChange) error {
	if s == nil || s.miniappProjector == nil {
		return nil
	}
	for _, change := range changes {
		switch change.Scope {
		case hmd.HmdScopeCentralizedProject:
			if err := s.miniappProjector.RefreshCentralizedProject(ctx, change.EntityID); err != nil {
				return databasef("apply hpd centralized project projection: %w", err)
			}
		case hmd.HmdScopeBuilding:
			if err := s.miniappProjector.RefreshBuilding(ctx, change.EntityID); err != nil {
				return databasef("apply hpd building projection: %w", err)
			}
		case hmd.HmdScopeRoomTypeCentralized:
			if err := s.miniappProjector.RefreshRoomTypeCentralized(ctx, change.EntityID); err != nil {
				return databasef("apply hpd room type centralized projection: %w", err)
			}
		case hmd.HmdScopeCentralizedRoom:
			if err := s.miniappProjector.RefreshCentralizedRoom(ctx, change.EntityID); err != nil {
				return databasef("apply hpd centralized room projection: %w", err)
			}
		case hmd.HmdScopeDecentralizedCommunity:
			if err := s.miniappProjector.RefreshDecentralizedCommunity(ctx, change.EntityID); err != nil {
				return databasef("apply hpd decentralized community projection: %w", err)
			}
		case hmd.HmdScopeDecentralizedRoom:
			if err := s.miniappProjector.RefreshDecentralizedRoom(ctx, change.EntityID); err != nil {
				return databasef("apply hpd decentralized room projection: %w", err)
			}
		}
	}
	return nil
}

func (s *Service) UpdateListingStatus(ctx context.Context, listingID bson.ObjectID, listingStatus hpdmodel.HpdListingStatus) error {
	if s == nil || s.listingRepo == nil {
		return nil
	}
	if err := s.listingRepo.UpdateStatus(ctx, listingID, listingStatus); err != nil {
		return databasef("update hpd listing status: %w", err)
	}
	return s.refreshListingAfterLifecycleUpdate(ctx, listingID)
}

func (s *Service) UpdateListingLifecycleFields(ctx context.Context, listingID bson.ObjectID, fields bson.M) error {
	if s == nil || s.listingRepo == nil {
		return nil
	}
	if err := s.listingRepo.UpdateLifecycleFields(ctx, listingID, fields); err != nil {
		return databasef("update hpd listing lifecycle fields: %w", err)
	}
	return s.refreshListingAfterLifecycleUpdate(ctx, listingID)
}

func (s *Service) refreshListingAfterLifecycleUpdate(ctx context.Context, listingID bson.ObjectID) error {
	if s.miniappProjector == nil {
		return nil
	}
	listing, err := s.listingRepo.FindByID(ctx, listingID)
	if err != nil {
		return databasef("refresh hpd listing after lifecycle update: %w", err)
	}
	if listing == nil {
		return databasef("refresh hpd listing after lifecycle update: listing not found")
	}
	if err := s.miniappProjector.RefreshByListing(ctx, listing); err != nil {
		return databasef("refresh hpd listing after lifecycle update: %w", err)
	}
	return nil
}
