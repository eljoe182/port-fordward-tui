package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"port-forward-tui/internal/domain"
)

func TestUpdateStoresCatalogOnLoadedMsg(t *testing.T) {
	m := NewModel(Dependencies{})

	result := CatalogResult{
		Contexts:   []string{"dev", "prod"},
		Namespaces: []string{"default", "cco"},
		Context:    "dev",
		Namespace:  "cco",
		Items: []CatalogItem{
			{ID: "service:admin", Label: "admin", RemotePort: 3000, PreferredLocalPort: 3001},
		},
	}

	next, _ := m.Update(catalogLoadedMsg{result: result})
	updated := next.(Model)

	if updated.contextName != "dev" || updated.namespace != "cco" {
		t.Fatalf("expected dev/cco, got %s/%s", updated.contextName, updated.namespace)
	}
	if len(updated.catalog) != 1 || updated.catalog[0].Label != "admin" {
		t.Fatalf("catalog not populated: %#v", updated.catalog)
	}
	if len(updated.contexts) != 2 || len(updated.namespaces) != 2 {
		t.Fatalf("expected contexts and namespaces persisted in model")
	}
	if updated.errMsg != "" {
		t.Fatalf("expected no error, got %q", updated.errMsg)
	}
}

func TestUpdateStoresErrorMessageOnCatalogErrorMsg(t *testing.T) {
	m := NewModel(Dependencies{})

	next, _ := m.Update(catalogErrorMsg{err: errors.New("boom")})
	updated := next.(Model)

	if updated.errMsg != "boom" {
		t.Fatalf("expected errMsg=boom, got %q", updated.errMsg)
	}
}

func TestFavoriteKeyPersistsCurrentCatalogItem(t *testing.T) {
	store := &fakeStore{cfg: domain.AppConfig{Targets: map[string]domain.TargetConfig{}}}
	m := NewModel(Dependencies{ConfigStore: store})
	m.catalog = []CatalogItem{{ID: "service:cco:admin", Type: "service", Namespace: "cco", Name: "admin", Label: "admin", RemotePort: 3000, PreferredLocalPort: 3001}}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	updated := next.(Model)
	if !updated.catalog[0].Favorite {
		t.Fatalf("expected favorite toggled in-memory")
	}
	if cmd == nil {
		t.Fatalf("expected persistence command")
	}
	_ = cmd()

	if !store.cfg.Targets["service:cco:admin"].Favorite {
		t.Fatalf("expected favorite persisted, got %+v", store.cfg.Targets["service:cco:admin"])
	}
}
