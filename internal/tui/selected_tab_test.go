package tui

import "testing"

func TestSelectingTargetAddsItToSelectedWithPreferredPort(t *testing.T) {
	m := NewModel(Dependencies{})
	m.catalog = []CatalogItem{{ID: "service:cco:admin", Label: "admin", PreferredLocalPort: 3001, RemotePort: 3000}}

	m.selectCurrentItem()

	if len(m.selected) != 1 || m.selected[0].LocalPort != 3001 {
		t.Fatalf("expected selected item with preferred local port, got %#v", m.selected)
	}
}
