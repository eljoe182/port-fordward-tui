package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"port-forward-tui/internal/app/catalog"
	"port-forward-tui/internal/domain"
)

func newModelWithCatalog(items ...CatalogItem) Model {
	m := NewModel(Dependencies{})
	m.catalog = items
	return m
}

func pressKey(t *testing.T, m Model, key string) Model {
	t.Helper()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	next, _ := m.Update(msg)
	return next.(Model)
}

func pressSpecial(t *testing.T, m Model, keyType tea.KeyType) Model {
	t.Helper()
	next, _ := m.Update(tea.KeyMsg{Type: keyType})
	return next.(Model)
}

func TestCursorMovesDownAndUpWithinCatalog(t *testing.T) {
	m := newModelWithCatalog(
		CatalogItem{ID: "a", Label: "a"},
		CatalogItem{ID: "b", Label: "b"},
		CatalogItem{ID: "c", Label: "c"},
	)

	m = pressKey(t, m, "j")
	if m.cursor != 1 {
		t.Fatalf("expected cursor 1 after j, got %d", m.cursor)
	}
	m = pressSpecial(t, m, tea.KeyDown)
	if m.cursor != 2 {
		t.Fatalf("expected cursor 2 after down, got %d", m.cursor)
	}
	m = pressKey(t, m, "j")
	if m.cursor != 2 {
		t.Fatalf("cursor must clamp at bottom, got %d", m.cursor)
	}
	m = pressKey(t, m, "k")
	if m.cursor != 1 {
		t.Fatalf("expected cursor 1 after k, got %d", m.cursor)
	}
	m = pressSpecial(t, m, tea.KeyUp)
	m = pressSpecial(t, m, tea.KeyUp)
	if m.cursor != 0 {
		t.Fatalf("cursor must clamp at top, got %d", m.cursor)
	}
}

func TestEnterAddsCurrentCatalogItemToSelected(t *testing.T) {
	m := newModelWithCatalog(
		CatalogItem{ID: "service:admin", Label: "admin", RemotePort: 3000, PreferredLocalPort: 3001},
	)

	m = pressSpecial(t, m, tea.KeyEnter)

	if len(m.selected) != 1 {
		t.Fatalf("expected 1 selected, got %d", len(m.selected))
	}
	if m.selected[0].LocalPort != 3001 {
		t.Fatalf("expected preferred local port 3001, got %d", m.selected[0].LocalPort)
	}
}

func TestEnterDoesNotDuplicateAlreadySelectedItem(t *testing.T) {
	m := newModelWithCatalog(
		CatalogItem{ID: "service:admin", Label: "admin", RemotePort: 3000, PreferredLocalPort: 3001},
	)

	m = pressSpecial(t, m, tea.KeyEnter)
	m = pressSpecial(t, m, tea.KeyEnter)

	if len(m.selected) != 1 {
		t.Fatalf("expected selected to stay at 1, got %d", len(m.selected))
	}
}

func TestContextAndNamespaceKeysPersistAndReloadCatalog(t *testing.T) {
	discovery := fakeDiscovery{
		currentContext: "dev",
		contexts:       []string{"dev", "prod"},
		namespaces:     []string{"default", "cco"},
		targets:        []domain.Target{{Name: "admin", Namespace: "default", Type: domain.TargetTypeService, RemotePort: 3000, Available: true}},
	}
	store := &fakeStore{cfg: domain.AppConfig{CurrentContext: "dev", CurrentNamespace: "default", Targets: map[string]domain.TargetConfig{}}}
	m := NewModel(Dependencies{Discovery: discovery, ConfigStore: store})
	m.contexts = []string{"dev", "prod"}
	m.namespaces = []string{"default", "cco"}
	m.contextName = "dev"
	m.namespace = "default"
	m.ctx = context.Background()

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	updated := next.(Model)
	if updated.modalKind != ModalContext {
		t.Fatalf("expected context modal, got %q", updated.modalKind)
	}
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated = next.(Model)
	next, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated = next.(Model)
	msg := cmd()
	next, _ = updated.Update(msg)
	updated = next.(Model)
	if store.cfg.CurrentContext != "prod" {
		t.Fatalf("expected current context persisted, got %q", store.cfg.CurrentContext)
	}

	next, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	updated = next.(Model)
	if updated.modalKind != ModalNamespace {
		t.Fatalf("expected namespace modal, got %q", updated.modalKind)
	}
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated = next.(Model)
	next, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated = next.(Model)
	msg = cmd()
	next, _ = updated.Update(msg)
	updated = next.(Model)
	if store.cfg.CurrentNamespace != "cco" {
		t.Fatalf("expected current namespace persisted, got %q", store.cfg.CurrentNamespace)
	}
}

func TestSearchModeAppliesQueryAndReloadsCatalog(t *testing.T) {
	discovery := fakeDiscovery{
		currentContext: "dev",
		contexts:       []string{"dev"},
		namespaces:     []string{"default"},
		targets: []domain.Target{
			{Name: "admin", Namespace: "default", Type: domain.TargetTypeService, RemotePort: 3000, Available: true},
			{Name: "redis", Namespace: "default", Type: domain.TargetTypePod, RemotePort: 6379, Available: true},
		},
	}
	store := &fakeStore{cfg: domain.AppConfig{CurrentContext: "dev", CurrentNamespace: "default", Targets: map[string]domain.TargetConfig{}}}
	m := NewModel(Dependencies{Discovery: discovery, ConfigStore: store}).WithContext(context.Background())
	m.contextName = "dev"
	m.namespace = "default"

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = next.(Model)
	if m.modalKind != ModalSearch {
		t.Fatalf("expected search modal, got %q", m.modalKind)
	}
	for _, r := range []rune("adm") {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(Model)
	}
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	msg := cmd()
	next, _ = m.Update(msg)
	m = next.(Model)

	if m.query != "adm" {
		t.Fatalf("expected query persisted, got %q", m.query)
	}
	if len(m.catalog) != 1 || m.catalog[0].Name != "admin" {
		t.Fatalf("expected filtered catalog for query, got %+v", m.catalog)
	}
}

func TestFilterAndSortKeysCycleModes(t *testing.T) {
	m := NewModel(Dependencies{})
	if m.filterMode != catalog.FilterAll || m.sortMode != catalog.SortSmart {
		t.Fatalf("unexpected initial modes: %s %s", m.filterMode, m.sortMode)
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	m = next.(Model)
	if m.modalKind != ModalFilter {
		t.Fatalf("expected filter modal, got %q", m.modalKind)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(Model)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd != nil {
		next, _ = m.Update(cmd())
		m = next.(Model)
	}
	if m.filterMode != catalog.FilterServices {
		t.Fatalf("expected services filter, got %s", m.filterMode)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	m = next.(Model)
	if m.modalKind != ModalSort {
		t.Fatalf("expected sort modal, got %q", m.modalKind)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(Model)
	next, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd != nil {
		next, _ = m.Update(cmd())
		m = next.(Model)
	}
	if m.sortMode != catalog.SortName {
		t.Fatalf("expected name sort, got %s", m.sortMode)
	}
}
