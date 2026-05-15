package listingprojection

import (
	"context"

	"house-manager/internal/domain/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	repohpd "house-manager/internal/repository/hpd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Service consumes HMD mutations and refreshes HPD listing read models.
// Concrete projection targets currently include the miniapp and publisher read models.
// Publish ownership relations live in package publishaccess instead of this projection service.
type Service struct {
	listingRepo        hpdListingRepository
	miniappProjector   miniappProjector
	publisherProjector publisherProjector
}

func NewService(
	hpdListingRepo *repohpd.ListingRepository,
	miniappProjector *MiniappProjector,
	publisherProjector *PublisherProjector,
) *Service {
	return &Service{
		listingRepo:        hpdListingRepo,
		miniappProjector:   miniappProjector,
		publisherProjector: publisherProjector,
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

type publisherProjector interface {
	RefreshByListing(ctx context.Context, listing *hpdmodel.HpdListing) error
	RefreshCentralizedRoom(ctx context.Context, roomID bson.ObjectID) error
	RefreshDecentralizedRoom(ctx context.Context, roomID bson.ObjectID) error
	RefreshCentralizedProject(ctx context.Context, projectID bson.ObjectID) error
	RefreshBuilding(ctx context.Context, buildingID bson.ObjectID) error
	RefreshRoomTypeCentralized(ctx context.Context, roomTypeID bson.ObjectID) error
	RefreshDecentralizedCommunity(ctx context.Context, decentralizedID bson.ObjectID) error
}

func (s *Service) Apply(ctx context.Context, changes []hmd.HmdChange) error {
	if s == nil {
		return nil
	}
	logProjectionInfo(ctx, "listingprojection.apply.start", "change_count", len(changes))
	for _, change := range changes {
		logProjectionInfo(ctx, "listingprojection.apply.change", "scope", change.Scope, "entity_id", change.EntityID.Hex())
		switch change.Scope {
		case hmd.HmdScopeCentralizedProject:
			if err := s.refreshCentralizedProject(ctx, change.EntityID); err != nil {
				logProjectionError(ctx, "listingprojection.apply.failed", "scope", change.Scope, "entity_id", change.EntityID.Hex(), "error", err)
				return databasef("apply hpd centralized project projection: %w", err)
			}
		case hmd.HmdScopeBuilding:
			if err := s.refreshBuilding(ctx, change.EntityID); err != nil {
				logProjectionError(ctx, "listingprojection.apply.failed", "scope", change.Scope, "entity_id", change.EntityID.Hex(), "error", err)
				return databasef("apply hpd building projection: %w", err)
			}
		case hmd.HmdScopeRoomTypeCentralized:
			if err := s.refreshRoomTypeCentralized(ctx, change.EntityID); err != nil {
				logProjectionError(ctx, "listingprojection.apply.failed", "scope", change.Scope, "entity_id", change.EntityID.Hex(), "error", err)
				return databasef("apply hpd room type centralized projection: %w", err)
			}
		case hmd.HmdScopeCentralizedRoom:
			if err := s.refreshCentralizedRoom(ctx, change.EntityID); err != nil {
				logProjectionError(ctx, "listingprojection.apply.failed", "scope", change.Scope, "entity_id", change.EntityID.Hex(), "error", err)
				return databasef("apply hpd centralized room projection: %w", err)
			}
		case hmd.HmdScopeDecentralizedCommunity:
			if err := s.refreshDecentralizedCommunity(ctx, change.EntityID); err != nil {
				logProjectionError(ctx, "listingprojection.apply.failed", "scope", change.Scope, "entity_id", change.EntityID.Hex(), "error", err)
				return databasef("apply hpd decentralized community projection: %w", err)
			}
		case hmd.HmdScopeDecentralizedRoom:
			if err := s.refreshDecentralizedRoom(ctx, change.EntityID); err != nil {
				logProjectionError(ctx, "listingprojection.apply.failed", "scope", change.Scope, "entity_id", change.EntityID.Hex(), "error", err)
				return databasef("apply hpd decentralized room projection: %w", err)
			}
		}
	}
	logProjectionInfo(ctx, "listingprojection.apply.success", "change_count", len(changes))
	return nil
}

func (s *Service) UpdateListingStatus(ctx context.Context, listingID bson.ObjectID, listingStatus hpdmodel.HpdListingStatus) error {
	if s == nil || s.listingRepo == nil {
		return nil
	}
	logProjectionInfo(ctx, "listingprojection.lifecycle.update_status.start", "listing_id", listingID.Hex(), "listing_status", listingStatus)
	if err := s.listingRepo.UpdateStatus(ctx, listingID, listingStatus); err != nil {
		logProjectionError(ctx, "listingprojection.lifecycle.update_status.failed", "listing_id", listingID.Hex(), "error", err)
		return databasef("update hpd listing status: %w", err)
	}
	err := s.refreshListingAfterLifecycleUpdate(ctx, listingID)
	if err != nil {
		logProjectionError(ctx, "listingprojection.lifecycle.update_status.failed", "listing_id", listingID.Hex(), "error", err)
		return err
	}
	logProjectionInfo(ctx, "listingprojection.lifecycle.update_status.success", "listing_id", listingID.Hex(), "listing_status", listingStatus)
	return nil
}

func (s *Service) UpdateListingLifecycleFields(ctx context.Context, listingID bson.ObjectID, fields bson.M) error {
	if s == nil || s.listingRepo == nil {
		return nil
	}
	logProjectionInfo(ctx, "listingprojection.lifecycle.update_fields.start", "listing_id", listingID.Hex(), "field_count", len(fields))
	if err := s.listingRepo.UpdateLifecycleFields(ctx, listingID, fields); err != nil {
		logProjectionError(ctx, "listingprojection.lifecycle.update_fields.failed", "listing_id", listingID.Hex(), "error", err)
		return databasef("update hpd listing lifecycle fields: %w", err)
	}
	err := s.refreshListingAfterLifecycleUpdate(ctx, listingID)
	if err != nil {
		logProjectionError(ctx, "listingprojection.lifecycle.update_fields.failed", "listing_id", listingID.Hex(), "error", err)
		return err
	}
	logProjectionInfo(ctx, "listingprojection.lifecycle.update_fields.success", "listing_id", listingID.Hex(), "field_count", len(fields))
	return nil
}

func (s *Service) refreshListingAfterLifecycleUpdate(ctx context.Context, listingID bson.ObjectID) error {
	if s.miniappProjector == nil && s.publisherProjector == nil {
		return nil
	}
	logProjectionInfo(ctx, "listingprojection.lifecycle.refresh.start", "listing_id", listingID.Hex())
	listing, err := s.listingRepo.FindByID(ctx, listingID)
	if err != nil {
		return databasef("refresh hpd listing after lifecycle update: %w", err)
	}
	if listing == nil {
		return databasef("refresh hpd listing after lifecycle update: listing not found")
	}
	if s.miniappProjector != nil {
		if err := s.miniappProjector.RefreshByListing(ctx, listing); err != nil {
			return databasef("refresh hpd listing after lifecycle update: %w", err)
		}
	}
	if s.publisherProjector != nil {
		if err := s.publisherProjector.RefreshByListing(ctx, listing); err != nil {
			return databasef("refresh hpd listing after lifecycle update: %w", err)
		}
	}
	logProjectionInfo(ctx, "listingprojection.lifecycle.refresh.success", "listing_id", listingID.Hex(), "source_type", listing.SourceType)
	return nil
}

func (s *Service) refreshCentralizedProject(ctx context.Context, projectID bson.ObjectID) error {
	if s.miniappProjector != nil {
		if err := s.miniappProjector.RefreshCentralizedProject(ctx, projectID); err != nil {
			return err
		}
	}
	if s.publisherProjector != nil {
		if err := s.publisherProjector.RefreshCentralizedProject(ctx, projectID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) refreshBuilding(ctx context.Context, buildingID bson.ObjectID) error {
	if s.miniappProjector != nil {
		if err := s.miniappProjector.RefreshBuilding(ctx, buildingID); err != nil {
			return err
		}
	}
	if s.publisherProjector != nil {
		if err := s.publisherProjector.RefreshBuilding(ctx, buildingID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) refreshRoomTypeCentralized(ctx context.Context, roomTypeID bson.ObjectID) error {
	if s.miniappProjector != nil {
		if err := s.miniappProjector.RefreshRoomTypeCentralized(ctx, roomTypeID); err != nil {
			return err
		}
	}
	if s.publisherProjector != nil {
		if err := s.publisherProjector.RefreshRoomTypeCentralized(ctx, roomTypeID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) refreshCentralizedRoom(ctx context.Context, roomID bson.ObjectID) error {
	if s.miniappProjector != nil {
		if err := s.miniappProjector.RefreshCentralizedRoom(ctx, roomID); err != nil {
			return err
		}
	}
	if s.publisherProjector != nil {
		if err := s.publisherProjector.RefreshCentralizedRoom(ctx, roomID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) refreshDecentralizedCommunity(ctx context.Context, decentralizedID bson.ObjectID) error {
	if s.miniappProjector != nil {
		if err := s.miniappProjector.RefreshDecentralizedCommunity(ctx, decentralizedID); err != nil {
			return err
		}
	}
	if s.publisherProjector != nil {
		if err := s.publisherProjector.RefreshDecentralizedCommunity(ctx, decentralizedID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) refreshDecentralizedRoom(ctx context.Context, roomID bson.ObjectID) error {
	if s.miniappProjector != nil {
		if err := s.miniappProjector.RefreshDecentralizedRoom(ctx, roomID); err != nil {
			return err
		}
	}
	if s.publisherProjector != nil {
		if err := s.publisherProjector.RefreshDecentralizedRoom(ctx, roomID); err != nil {
			return err
		}
	}
	return nil
}
