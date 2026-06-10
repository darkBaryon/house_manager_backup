package listingprojection

import (
	"context"
	"log/slog"

	"house-manager/internal/domain/hmd"
	hpdmodel "house-manager/internal/model/hpd"
	repohpd "house-manager/internal/repository/hpd"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Service consumes HMD mutations and refreshes HPD listing read models.
// Concrete projection targets currently include the miniapp, publisher and admin read models.
// Publish ownership relations live in package publishaccess instead of this projection service.
type Service struct {
	listingRepo hpdListingRepository
	projectors  []projector
}

// projector 是读模型投影器的窄接口：Refresh 按变更回执幂等刷新对应实体
// （scope 的处理知识内聚在各 projector 的 dispatch 中），RefreshByListing
// 服务于 lifecycle 路径（按 listing 对象刷新）。
type projector interface {
	Refresh(ctx context.Context, change hmd.HmdChange) error
	RefreshByListing(ctx context.Context, listing *hpdmodel.HpdListing) error
}

// NewService 保持具名参数签名（wire 零改动）；投影顺序在此集中声明：
// miniapp → publisher → admin。新增 projector 只需实现 projector 接口
// 并在切片中登记。
func NewService(
	hpdListingRepo *repohpd.ListingRepository,
	miniappProjector *MiniappProjector,
	publisherProjector *PublisherProjector,
	adminProjector *AdminProjector,
) *Service {
	return &Service{
		listingRepo: hpdListingRepo,
		projectors:  []projector{miniappProjector, publisherProjector, adminProjector},
	}
}

func (s *Service) Apply(ctx context.Context, changes []hmd.HmdChange) error {
	if s == nil {
		return nil
	}
	slog.InfoContext(ctx, "listingprojection.apply.start", "change_count", len(changes))
	for _, change := range changes {
		slog.InfoContext(ctx, "listingprojection.apply.change", "scope", change.Scope, "entity_id", change.EntityID.Hex())
		for _, p := range s.projectors {
			if p == nil {
				continue
			}
			if err := p.Refresh(ctx, change); err != nil {
				slog.ErrorContext(ctx, "listingprojection.apply.failed", "scope", change.Scope, "entity_id", change.EntityID.Hex(), "error", err)
				return databasef("apply hpd %s projection: %w", change.Scope, err)
			}
		}
	}
	slog.InfoContext(ctx, "listingprojection.apply.success", "change_count", len(changes))
	return nil
}

func (s *Service) UpdateListingStatus(ctx context.Context, listingID bson.ObjectID, listingStatus hpdmodel.HpdListingStatus) error {
	if s == nil || s.listingRepo == nil {
		return nil
	}
	slog.InfoContext(ctx, "listingprojection.lifecycle.update_status.start", "listing_id", listingID.Hex(), "listing_status", listingStatus)
	if err := s.listingRepo.UpdateStatus(ctx, listingID, listingStatus); err != nil {
		slog.ErrorContext(ctx, "listingprojection.lifecycle.update_status.failed", "listing_id", listingID.Hex(), "error", err)
		return databasef("update hpd listing status: %w", err)
	}
	err := s.refreshListingAfterLifecycleUpdate(ctx, listingID)
	if err != nil {
		slog.ErrorContext(ctx, "listingprojection.lifecycle.update_status.failed", "listing_id", listingID.Hex(), "error", err)
		return err
	}
	slog.InfoContext(ctx, "listingprojection.lifecycle.update_status.success", "listing_id", listingID.Hex(), "listing_status", listingStatus)
	return nil
}

func (s *Service) UpdateListingLifecycleFields(ctx context.Context, listingID bson.ObjectID, fields bson.M) error {
	if s == nil || s.listingRepo == nil {
		return nil
	}
	slog.InfoContext(ctx, "listingprojection.lifecycle.update_fields.start", "listing_id", listingID.Hex(), "field_count", len(fields))
	if err := s.listingRepo.UpdateLifecycleFields(ctx, listingID, fields); err != nil {
		slog.ErrorContext(ctx, "listingprojection.lifecycle.update_fields.failed", "listing_id", listingID.Hex(), "error", err)
		return databasef("update hpd listing lifecycle fields: %w", err)
	}
	err := s.refreshListingAfterLifecycleUpdate(ctx, listingID)
	if err != nil {
		slog.ErrorContext(ctx, "listingprojection.lifecycle.update_fields.failed", "listing_id", listingID.Hex(), "error", err)
		return err
	}
	slog.InfoContext(ctx, "listingprojection.lifecycle.update_fields.success", "listing_id", listingID.Hex(), "field_count", len(fields))
	return nil
}

func (s *Service) refreshListingAfterLifecycleUpdate(ctx context.Context, listingID bson.ObjectID) error {
	hasProjector := false
	for _, p := range s.projectors {
		if p != nil {
			hasProjector = true
			break
		}
	}
	if !hasProjector {
		return nil
	}
	slog.InfoContext(ctx, "listingprojection.lifecycle.refresh.start", "listing_id", listingID.Hex())
	listing, err := s.listingRepo.FindByID(ctx, listingID)
	if err != nil {
		return databasef("refresh hpd listing after lifecycle update: %w", err)
	}
	if listing == nil {
		return databasef("refresh hpd listing after lifecycle update: listing not found")
	}
	for _, p := range s.projectors {
		if p == nil {
			continue
		}
		if err := p.RefreshByListing(ctx, listing); err != nil {
			return databasef("refresh hpd listing after lifecycle update: %w", err)
		}
	}
	slog.InfoContext(ctx, "listingprojection.lifecycle.refresh.success", "listing_id", listingID.Hex(), "source_type", listing.SourceType)
	return nil
}
