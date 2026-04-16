package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"cco-port-forward-tui/internal/app/catalog"
	"cco-port-forward-tui/internal/domain"
)

type CatalogResult struct {
	Context   string
	Namespace string
	Items     []CatalogItem
}

type catalogLoadedMsg struct{ result CatalogResult }

type catalogErrorMsg struct{ err error }

func LoadCatalog(ctx context.Context, deps Dependencies) (CatalogResult, error) {
	cfg, err := deps.ConfigStore.Load()
	if err != nil {
		return CatalogResult{}, fmt.Errorf("load config: %w", err)
	}

	contextName := cfg.CurrentContext
	if contextName == "" {
		contextName, err = deps.Discovery.CurrentContext(ctx)
		if err != nil {
			return CatalogResult{}, fmt.Errorf("resolve current context: %w", err)
		}
	}

	namespace := cfg.CurrentNamespace
	if namespace == "" {
		namespaces, err := deps.Discovery.ListNamespaces(ctx, contextName)
		if err != nil {
			return CatalogResult{}, fmt.Errorf("list namespaces: %w", err)
		}
		namespace = pickNamespace(namespaces)
	}

	targets, err := deps.Discovery.ListTargets(ctx, contextName, namespace)
	if err != nil {
		return CatalogResult{}, fmt.Errorf("list targets: %w", err)
	}

	merged := catalog.MergeTargets(targets, cfg.Targets)
	ranked := catalog.RankSmart(merged, time.Now(), "")

	items := make([]CatalogItem, 0, len(ranked))
	for _, target := range ranked {
		items = append(items, CatalogItem{
			ID:                 catalog.TargetKey(target),
			Label:              targetLabel(target),
			RemotePort:         target.RemotePort,
			PreferredLocalPort: choosePreferredPort(target),
		})
	}

	return CatalogResult{
		Context:   contextName,
		Namespace: namespace,
		Items:     items,
	}, nil
}

func loadCatalogCmd(ctx context.Context, deps Dependencies) tea.Cmd {
	return func() tea.Msg {
		result, err := LoadCatalog(ctx, deps)
		if err != nil {
			return catalogErrorMsg{err: err}
		}
		return catalogLoadedMsg{result: result}
	}
}

func pickNamespace(namespaces []string) string {
	for _, ns := range namespaces {
		if ns == "default" {
			return ns
		}
	}
	if len(namespaces) > 0 {
		return namespaces[0]
	}
	return "default"
}

func targetLabel(target domain.Target) string {
	if target.Alias != "" {
		return target.Alias
	}
	return target.Name
}

func choosePreferredPort(target domain.Target) int {
	if target.PreferredLocalPort != 0 {
		return target.PreferredLocalPort
	}
	return target.RemotePort
}
