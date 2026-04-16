package tui

import (
	"errors"
	"testing"
)

func TestUpdateStoresCatalogOnLoadedMsg(t *testing.T) {
	m := NewModel(Dependencies{})

	result := CatalogResult{
		Context:   "dev",
		Namespace: "cco",
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
