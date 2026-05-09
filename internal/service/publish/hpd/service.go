package hpd

import (
	"context"

	"house-manager/internal/service/publish/hmd"
)

// Service 是展示层数据同步的预留入口。
//
// 当前阶段只实现 HMD 写入链路，HPD 集合和投影规则还没定稿，所以 Apply 保持 no-op。
// 等 HPD 模型落地后，这里会把 HMD changes 交给 projector 或 outbox worker 复用的投影逻辑。
type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Apply(ctx context.Context, changes []hmd.HmdChange) error {
	return nil
}
