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
	Contexts   []string
	Namespaces []string
	Context    string
	Namespace  string
	Query      string
	Filter     string
	Sort       string
	Items      []CatalogItem
}

type catalogLoadedMsg struct{ result CatalogResult }

type catalogErrorMsg struct{ err error }

func LoadCatalog(ctx context.Context, deps Dependencies, opts catalog.LoadOptions) (CatalogResult, error) {
	opts = opts.WithDefaults()
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

	contexts, err := deps.Discovery.ListContexts(ctx)
	if err != nil {
		return CatalogResult{}, fmt.Errorf("list contexts: %w", err)
	}

	namespace := cfg.CurrentNamespace
	namespaces, err := deps.Discovery.ListNamespaces(ctx, contextName)
	if err != nil {
		return CatalogResult{}, fmt.Errorf("list namespaces: %w", err)
	}
	if namespace == "" || !contains(namespaces, namespace) {
		namespace = pickNamespace(namespaces)
	}

	targets, err := deps.Discovery.ListTargets(ctx, contextName, namespace)
	if err != nil {
		return CatalogResult{}, fmt.Errorf("list targets: %w", err)
	}

	merged := catalog.MergeTargets(targets, cfg.Targets)
	filtered := catalog.ApplyFilters(merged, opts)
	ranked := catalog.SortTargets(filtered, time.Now(), opts)

	items := make([]CatalogItem, 0, len(ranked))
	for _, target := range ranked {
		items = append(items, CatalogItem{
			Type:               string(target.Type),
			Namespace:          target.Namespace,
			Name:               target.Name,
			ID:                 catalog.TargetKey(target),
			Label:              targetLabel(target),
			RemotePort:         target.RemotePort,
			PreferredLocalPort: choosePreferredPort(target),
			Favorite:           target.Favorite,
			Available:          target.Available,
		})
	}

	return CatalogResult{
		Contexts:   contexts,
		Namespaces: namespaces,
		Context:    contextName,
		Namespace:  namespace,
		Query:      opts.Query,
		Filter:     string(opts.Filter),
		Sort:       string(opts.Sort),
		Items:      items,
	}, nil
}

func loadCatalogCmd(ctx context.Context, deps Dependencies, opts catalog.LoadOptions) tea.Cmd {
	return func() tea.Msg {
		result, err := LoadCatalog(ctx, deps, opts)
		if err != nil {
			return catalogErrorMsg{err: err}
		}
		return catalogLoadedMsg{result: result}
	}
}

func saveConfigAndReloadCmd(ctx context.Context, deps Dependencies, opts catalog.LoadOptions, mutate func(*domain.AppConfig)) tea.Cmd {
	return func() tea.Msg {
		cfg, err := deps.ConfigStore.Load()
		if err != nil {
			return catalogErrorMsg{err: fmt.Errorf("load config: %w", err)}
		}
		mutate(&cfg)
		if err := deps.ConfigStore.Save(cfg); err != nil {
			return catalogErrorMsg{err: fmt.Errorf("save config: %w", err)}
		}
		result, err := LoadCatalog(ctx, deps, opts)
		if err != nil {
			return catalogErrorMsg{err: err}
		}
		return catalogLoadedMsg{result: result}
	}
}

func saveConfigCmd(deps Dependencies, mutate func(*domain.AppConfig)) tea.Cmd {
	return func() tea.Msg {
		cfg, err := deps.ConfigStore.Load()
		if err != nil {
			return catalogErrorMsg{err: fmt.Errorf("load config: %w", err)}
		}
		mutate(&cfg)
		if err := deps.ConfigStore.Save(cfg); err != nil {
			return catalogErrorMsg{err: fmt.Errorf("save config: %w", err)}
		}
		return nil
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

func contains(items []string, wanted string) bool {
	for _, item := range items {
		if item == wanted {
			return true
		}
	}
	return false
}
