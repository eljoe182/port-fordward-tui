package tui

import (
	"strings"
	"testing"
)

func TestViewRendersWorkspaceWithCatalogAndPanelTabs(t *testing.T) {
	m := NewModel(Dependencies{})
	m.height = 20
	m.contextName = "dev"
	m.namespace = "cco"
	m.catalog = []CatalogItem{{ID: "service:cco:admin", Type: "service", Namespace: "cco", Name: "admin", Label: "admin", RemotePort: 3000, PreferredLocalPort: 3001, Favorite: true, Available: true}}
	m.selected = []SelectedItem{{TargetID: "service:cco:admin", Label: "admin", LocalPort: 3001, RemotePort: 3000}}

	view := m.View()

	for _, snippet := range []string{"Catalog", "Panel", "Selected", "Running", "admin", "ctx=dev", "ns=cco"} {
		if !strings.Contains(view, snippet) {
			t.Fatalf("expected view to contain %q, got:\n%s", snippet, view)
		}
	}
}

func TestViewShowsRunningPanelContentWhenRunningTabActive(t *testing.T) {
	m := NewModel(Dependencies{})
	m.height = 20
	m.activeTab = TabRunning
	m.running = []RunningItem{{TargetID: "service:cco:admin", Label: "admin", LocalPort: 3001, RemotePort: 3000, Status: StatusFailed, Err: "local port unavailable — edit the local port and retry"}}

	view := m.View()
	if !strings.Contains(view, "press R to retry") {
		t.Fatalf("expected retry hint in running panel, got:\n%s", view)
	}
}

func TestViewKeepsHeaderVisibleWhenCatalogIsLarge(t *testing.T) {
	m := NewModel(Dependencies{})
	m.contextName = "dev"
	m.namespace = "default"
	for i := 0; i < 40; i++ {
		m.catalog = append(m.catalog, CatalogItem{ID: "service:default:item", Type: "service", Namespace: "default", Name: "item", Label: "item", RemotePort: 3000, PreferredLocalPort: 3000, Available: true})
	}
	m.cursor = 30

	view := m.View()
	if !strings.Contains(view, "ctx=dev") {
		t.Fatalf("expected header to remain visible, got:\n%s", view)
	}
	if !strings.Contains(view, "↑") || !strings.Contains(view, "↓") {
		t.Fatalf("expected clipped catalog indicators, got:\n%s", view)
	}
}

func TestViewRendersSelectorModal(t *testing.T) {
	m := NewModel(Dependencies{})
	m.contexts = []string{"dev", "prod"}
	m.contextName = "dev"
	m = m.openContextModal()

	view := m.View()
	if !strings.Contains(view, "Select context") || !strings.Contains(view, "prod") {
		t.Fatalf("expected context modal in view, got:\n%s", view)
	}
}
