package ports

import (
	"context"

	"cco-port-forward-tui/internal/domain"
)

type KubernetesDiscovery interface {
	ListContexts(ctx context.Context) ([]string, error)
	ListNamespaces(ctx context.Context, contextName string) ([]string, error)
	ListTargets(ctx context.Context, contextName, namespace string) ([]domain.Target, error)
}
