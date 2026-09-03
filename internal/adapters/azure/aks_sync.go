package azure

import (
	"context"
	"fmt"
	"strings"

	"port-forward-tui/internal/ports"
)

const providerName = "Azure AKS"

// ExecRunner runs an external command and returns combined-capable stdout.
type ExecRunner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

// AKSSyncer syncs AKS cluster credentials into kubeconfig via the Azure CLI.
type AKSSyncer struct {
	exec ExecRunner
}

// NewAKSSyncer constructs an AKSSyncer that invokes az through exec.
func NewAKSSyncer(exec ExecRunner) AKSSyncer {
	return AKSSyncer{exec: exec}
}

// Provider returns the human-readable cloud provider name.
func (AKSSyncer) Provider() string { return providerName }

// Sync lists subscriptions and AKS clusters, then runs get-credentials for each.
// Setup failures (az missing, not logged in, account list) return an error.
// Per-cluster get-credentials failures are collected in the result.
func (s AKSSyncer) Sync(ctx context.Context) (ports.CredentialSyncResult, error) {
	result := ports.CredentialSyncResult{Provider: providerName}

	out, err := s.exec.Run(ctx, "az", "account", "list", "--query", "[].id", "-o", "tsv")
	if err != nil {
		return result, fmt.Errorf("list Azure subscriptions: %w", err)
	}

	subs := nonEmptyLines(out)
	if len(subs) == 0 {
		return result, nil
	}

	for _, sub := range subs {
		clusters, err := s.listClusters(ctx, sub)
		if err != nil {
			return result, fmt.Errorf("list AKS clusters for subscription %s: %w", sub, err)
		}
		for _, cluster := range clusters {
			if err := s.getCredentials(ctx, sub, cluster); err != nil {
				result.Failed = append(result.Failed, ports.CredentialSyncFailure{
					Name:   fmt.Sprintf("%s/%s", cluster.resourceGroup, cluster.name),
					Detail: err.Error(),
				})
				continue
			}
			result.Synced++
		}
	}

	return result, nil
}

type aksCluster struct {
	name          string
	resourceGroup string
}

func (s AKSSyncer) listClusters(ctx context.Context, subscription string) ([]aksCluster, error) {
	out, err := s.exec.Run(ctx,
		"az", "aks", "list",
		"--subscription", subscription,
		"--query", "[].[name, resourceGroup]",
		"-o", "tsv",
	)
	if err != nil {
		return nil, err
	}

	lines := nonEmptyLines(out)
	clusters := make([]aksCluster, 0, len(lines))
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		clusters = append(clusters, aksCluster{name: parts[0], resourceGroup: parts[1]})
	}
	return clusters, nil
}

func (s AKSSyncer) getCredentials(ctx context.Context, subscription string, cluster aksCluster) error {
	_, err := s.exec.Run(ctx,
		"az", "aks", "get-credentials",
		"--resource-group", cluster.resourceGroup,
		"--name", cluster.name,
		"--subscription", subscription,
		"--overwrite-existing",
	)
	return err
}

func nonEmptyLines(s string) []string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil
	}
	raw := strings.Split(trimmed, "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
