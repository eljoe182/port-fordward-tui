package tui

import "testing"

func TestSelectingTargetAddsItToSelectedWithPreferredPort(t *testing.T) {
	m := NewModel(Dependencies{})
	m.catalog = []CatalogItem{{
		Context:            "dev",
		Namespace:          "cco",
		Type:               "service",
		ID:                 "service:cco:admin",
		Label:              "admin",
		PreferredLocalPort: 3001,
		RemotePort:         3000,
	}}

	m.selectCurrentItem()

	if len(m.selected) != 1 || m.selected[0].LocalPort != 3001 {
		t.Fatalf("expected selected item with preferred local port, got %#v", m.selected)
	}
	selected := m.selected[0]
	if selected.Context != "dev" || selected.Namespace != "cco" || selected.Type != "service" {
		t.Fatalf("expected selection provenance to be captured, got %+v", selected)
	}
}

func TestSelectingHomonymousTargetFromDifferentContextAddsSecondSelection(t *testing.T) {
	m := NewModel(Dependencies{})
	m.catalog = []CatalogItem{{
		Context: "dev", Namespace: "cco", Type: "service", ID: "service:cco:admin", Label: "admin", PreferredLocalPort: 3001,
	}}
	m.selectCurrentItem()

	m.catalog[0].Context = "prod"
	m.selectCurrentItem()

	if len(m.selected) != 2 {
		t.Fatalf("expected one selection per context, got %#v", m.selected)
	}
	if m.selected[0].Context != "dev" || m.selected[1].Context != "prod" {
		t.Fatalf("expected dev and prod selections, got %#v", m.selected)
	}
}
