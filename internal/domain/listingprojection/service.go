package listingprojection

import (
	"context"
	"log/slog"

	"house-manager/internal/domain/hmd"
)

// Service consumes HMD mutations and refreshes HPD listing read models.
// Concrete projection targets currently include the miniapp, publisher and admin read models.
// Publish ownership relations live in package publishaccess instead of this projection service.
type Service struct {
	projectors []projector
}

// projector 是读模型投影器的窄接口：Refresh 按变更回执幂等刷新对应实体。
// scope 的处理知识内聚在各 projector 的 dispatch 中。
type projector interface {
	Refresh(ctx context.Context, change hmd.HmdChange) error
}

// NewService 集中声明投影顺序：miniapp → publisher → admin。
// 新增 projector 只需实现 projector 接口并在切片中登记。
func NewService(
	miniappProjector *MiniappProjector,
	publisherProjector *PublisherProjector,
	adminProjector *AdminProjector,
) *Service {
	return &Service{
		projectors: []projector{miniappProjector, publisherProjector, adminProjector},
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
