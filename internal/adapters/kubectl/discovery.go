package kubectl

import (
	"context"
	"strings"

	"cco-port-forward-tui/internal/domain"
)

type ExecRunner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

type DiscoveryClient struct {
	exec ExecRunner
}

func NewDiscoveryClient(exec ExecRunner) DiscoveryClient {
	return DiscoveryClient{exec: exec}
}

func (c DiscoveryClient) ListContexts(ctx context.Context) ([]string, error) {
	out, err := c.exec.Run(ctx, "kubectl", "config", "get-contexts", "-o", "name")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}
	return lines, nil
}

func (c DiscoveryClient) ListNamespaces(ctx context.Context, contextName string) ([]string, error) {
	out, err := c.exec.Run(ctx, "kubectl",
		"--context", contextName,
		"get", "namespaces",
		"-o", "jsonpath={.items[*].metadata.name}",
	)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return []string{}, nil
	}
	return strings.Fields(trimmed), nil
}

func (c DiscoveryClient) ListTargets(ctx context.Context, contextName, namespace string) ([]domain.Target, error) {
	return []domain.Target{}, nil
}
