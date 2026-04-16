package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"port-forward-tui/internal/app/catalog"
	"port-forward-tui/internal/domain"
)

func (m Model) cycleContext() (Model, teaCmd) {
	if len(m.contexts) == 0 || m.deps.ConfigStore == nil {
		return m, nil
	}
	idx := indexOf(m.contexts, m.contextName)
	idx = (idx + 1) % len(m.contexts)
	nextContext := m.contexts[idx]
	m.errMsg = ""
	return m, saveConfigAndReloadCmd(m.ctx, m.deps, m.loadOptions(), func(cfg *domain.AppConfig) {
		cfg.CurrentContext = nextContext
		cfg.CurrentNamespace = ""
	})
}

func (m Model) cycleNamespace() (Model, teaCmd) {
	if len(m.namespaces) == 0 || m.deps.ConfigStore == nil {
		return m, nil
	}
	idx := indexOf(m.namespaces, m.namespace)
	idx = (idx + 1) % len(m.namespaces)
	nextNamespace := m.namespaces[idx]
	m.errMsg = ""
	return m, saveConfigAndReloadCmd(m.ctx, m.deps, m.loadOptions(), func(cfg *domain.AppConfig) {
		cfg.CurrentContext = m.contextName
		cfg.CurrentNamespace = nextNamespace
	})
}

func (m Model) refreshCatalog() teaCmd {
	if m.deps.ConfigStore == nil {
		return nil
	}
	return saveConfigAndReloadCmd(m.ctx, m.deps, m.loadOptions(), func(cfg *domain.AppConfig) {
		cfg.CurrentContext = m.contextName
		cfg.CurrentNamespace = m.namespace
	})
}

func (m Model) cycleFilterMode() (Model, teaCmd) {
	modes := []catalog.FilterMode{catalog.FilterAll, catalog.FilterServices, catalog.FilterPods, catalog.FilterFavorites}
	current := 0
	for i, mode := range modes {
		if m.filterMode == mode {
			current = i
			break
		}
	}
	m.filterMode = modes[(current+1)%len(modes)]
	return m, loadCatalogCmd(m.ctx, m.deps, m.loadOptions())
}

func (m Model) cycleSortMode() (Model, teaCmd) {
	modes := []catalog.SortMode{catalog.SortSmart, catalog.SortName, catalog.SortRecent, catalog.SortFavorites, catalog.SortType}
	current := 0
	for i, mode := range modes {
		if m.sortMode == mode {
			current = i
			break
		}
	}
	m.sortMode = modes[(current+1)%len(modes)]
	return m, loadCatalogCmd(m.ctx, m.deps, m.loadOptions())
}

func (m Model) toggleFavoriteCurrentItem() teaCmd {
	if m.deps.ConfigStore == nil || len(m.catalog) == 0 || m.cursor < 0 || m.cursor >= len(m.catalog) {
		return nil
	}
	item := m.catalog[m.cursor]
	return saveConfigCmd(m.deps, func(cfg *domain.AppConfig) {
		ensureTargetsMap(cfg)
		entry := configFromCatalogItem(item)
		if existing, ok := cfg.Targets[item.ID]; ok {
			entry = mergeConfig(entry, existing)
		}
		entry.Favorite = item.Favorite
		cfg.Targets[item.ID] = entry
	})
}

func (m Model) persistSelectedPort(item SelectedItem) teaCmd {
	if m.deps.ConfigStore == nil {
		return nil
	}
	return saveConfigCmd(m.deps, func(cfg *domain.AppConfig) {
		ensureTargetsMap(cfg)
		entry := configFromSelectedItem(item)
		if existing, ok := cfg.Targets[item.TargetID]; ok {
			entry = mergeConfig(entry, existing)
		}
		entry.PreferredLocalPort = item.LocalPort
		cfg.Targets[item.TargetID] = entry
	})
}

func (m Model) persistRecentSelections(items []SelectedItem) teaCmd {
	if m.deps.ConfigStore == nil || len(items) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return saveConfigCmd(m.deps, func(cfg *domain.AppConfig) {
		ensureTargetsMap(cfg)
		cfg.CurrentContext = m.contextName
		cfg.CurrentNamespace = m.namespace
		for _, item := range items {
			entry := configFromSelectedItem(item)
			if existing, ok := cfg.Targets[item.TargetID]; ok {
				entry = mergeConfig(entry, existing)
			}
			entry.LastUsedAt = now
			entry.PreferredLocalPort = item.LocalPort
			cfg.Targets[item.TargetID] = entry
		}
	})
}

type teaCmd = tea.Cmd

func ensureTargetsMap(cfg *domain.AppConfig) {
	if cfg.Targets == nil {
		cfg.Targets = map[string]domain.TargetConfig{}
	}
}

func configFromCatalogItem(item CatalogItem) domain.TargetConfig {
	ref, _ := domain.ParseTargetKey(item.ID)
	alias := ""
	if item.Label != "" && item.Label != item.Name {
		alias = item.Label
	}
	return domain.TargetConfig{
		Type:               domain.TargetType(item.Type),
		Namespace:          firstNonEmpty(item.Namespace, ref.Namespace),
		Name:               firstNonEmpty(item.Name, ref.Name),
		Alias:              alias,
		PreferredLocalPort: item.PreferredLocalPort,
		Favorite:           item.Favorite,
		RemotePort:         item.RemotePort,
	}
}

func configFromSelectedItem(item SelectedItem) domain.TargetConfig {
	ref, _ := domain.ParseTargetKey(item.TargetID)
	alias := ""
	if item.Label != "" && item.Label != ref.Name {
		alias = item.Label
	}
	return domain.TargetConfig{
		Type:               ref.Type,
		Namespace:          ref.Namespace,
		Name:               ref.Name,
		Alias:              alias,
		PreferredLocalPort: item.LocalPort,
		RemotePort:         item.RemotePort,
	}
}

func mergeConfig(base, existing domain.TargetConfig) domain.TargetConfig {
	if base.Type == "" {
		base.Type = existing.Type
	}
	if base.Namespace == "" {
		base.Namespace = existing.Namespace
	}
	if base.Name == "" {
		base.Name = existing.Name
	}
	if base.Alias == "" {
		base.Alias = existing.Alias
	}
	if base.PreferredLocalPort == 0 {
		base.PreferredLocalPort = existing.PreferredLocalPort
	}
	if base.RemotePort == 0 {
		base.RemotePort = existing.RemotePort
	}
	base.Favorite = base.Favorite || existing.Favorite
	if base.LastUsedAt.IsZero() {
		base.LastUsedAt = existing.LastUsedAt
	}
	return base
}

func indexOf(items []string, current string) int {
	for i, item := range items {
		if item == current {
			return i
		}
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
