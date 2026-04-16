package catalog

import (
	"context"
	"time"

	"cco-port-forward-tui/internal/domain"
	"cco-port-forward-tui/internal/ports"
)

type Service struct {
	discovery ports.KubernetesDiscovery
}

func NewService(discovery ports.KubernetesDiscovery) Service {
	return Service{discovery: discovery}
}

func (s Service) Load(ctx context.Context, contextName, namespace string, configs map[string]domain.TargetConfig, query string) ([]domain.Target, error) {
	discovered, err := s.discovery.ListTargets(ctx, contextName, namespace)
	if err != nil {
		return nil, err
	}
	merged := MergeTargets(discovered, configs)
	return RankSmart(merged, time.Now(), query), nil
}
