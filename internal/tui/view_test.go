package tui

import (
	"strings"
	"testing"
)

func TestViewRendersWorkspaceWithCatalogAndPanelTabs(t *testing.T) {
	m := NewModel(Dependencies{})
	m.height = 20
	m.width = 140
	m.contextName = "dev"
	m.namespace = "cco"
	m.catalog = []CatalogItem{{ID: "service:cco:admin", Type: "service", Namespace: "cco", Name: "admin", Label: "admin", RemotePort: 3000, PreferredLocalPort: 3001, Favorite: true, Available: true}}
	m.selected = []SelectedItem{{Context: "dev", Namespace: "cco", TargetID: "service:cco:admin", Label: "admin", LocalPort: 3001, RemotePort: 3000}}

	view := m.View()

	for _, snippet := range []string{"Catalog", "Panel", "Selected", "Running", "admin", "ctx=dev", "ns=cco", "[dev/cco]", "x remove/stop", "? extra", "Name | Type | NS | Port"} {
		if !strings.Contains(view, snippet) {
			t.Fatalf("expected view to contain %q, got:\n%s", snippet, view)
		}
	}
	if strings.Contains(view, "ctrl+r aks sync") {
		t.Fatalf("expected ctrl+r to live in Extra modal, not the lean footer, got:\n%s", view)
	}
}

func TestViewShowsRunningPanelContentWhenRunningTabActive(t *testing.T) {
	m := NewModel(Dependencies{})
	m.height = 20
	m.width = 140
	m.activeTab = TabRunning
	m.running = []RunningItem{{Context: "prod", Namespace: "cco", TargetID: "service:cco:admin", Label: "admin", LocalPort: 3001, RemotePort: 3000, Status: StatusFailed, Err: "local port unavailable — edit the local port and retry"}}

	view := m.View()
	for _, snippet := range []string{"press R to retry", "[prod/cco]", "R retry"} {
		if !strings.Contains(view, snippet) {
			t.Fatalf("expected %q in running panel, got:\n%s", snippet, view)
		}
	}
}

func TestViewKeepsHeaderVisibleWhenCatalogIsLarge(t *testing.T) {
	m := NewModel(Dependencies{})
	m.width = 140
	m.height = 24
	m.contextName = "dev"
	m.namespace = "default"
	m.filterMode = "all"
	for i := 0; i < 40; i++ {
		m.catalog = append(m.catalog, CatalogItem{ID: "service:default:item", Type: "service", Namespace: "default", Name: "item", Label: "item", RemotePort: 3000, PreferredLocalPort: 3000, Available: true})
	}
	m.cursor = 30

	view := m.View()
	if !strings.Contains(view, "ctx=dev") {
		t.Fatalf("expected footer context to remain visible, got:\n%s", view)
	}
	if !strings.Contains(view, "filter=") {
		t.Fatalf("expected header filter meta, got:\n%s", view)
	}
	if !strings.Contains(view, "↑") || !strings.Contains(view, "↓") {
		t.Fatalf("expected clipped catalog indicators, got:\n%s", view)
	}
}

func TestViewRendersSelectorModal(t *testing.T) {
	m := NewModel(Dependencies{})
	m.width = 140
	m.height = 30
	m.contexts = []string{"dev", "prod"}
	m.contextName = "dev"
	m = m.openContextModal()

	view := m.View()
	if !strings.Contains(view, "Select context") || !strings.Contains(view, "prod") {
		t.Fatalf("expected context modal in view, got:\n%s", view)
	}
	if !strings.Contains(view, "Enter select") {
		t.Fatalf("expected modal navigation hints, got:\n%s", view)
	}
}

func TestViewRendersExtraModal(t *testing.T) {
	m := NewModel(Dependencies{})
	m.width = 140
	m.height = 30
	m = m.openExtraModal()

	view := m.View()
	for _, snippet := range []string{"Extra", "ctrl+r aks sync", "r refresh", "Esc close"} {
		if !strings.Contains(view, snippet) {
			t.Fatalf("expected Extra modal to contain %q, got:\n%s", snippet, view)
		}
	}
}

func TestViewCentersSelectorModalOverWorkspace(t *testing.T) {
	m := NewModel(Dependencies{})
	m.width = 140
	m.height = 40
	m.contexts = []string{"dev", "prod", "staging"}
	m.contextName = "dev"
	m.catalog = []CatalogItem{{ID: "service:default:admin", Type: "service", Namespace: "default", Label: "admin", RemotePort: 3000, Available: true}}
	m = m.openContextModal()

	view := m.View()
	if strings.Contains(view, "\r") {
		t.Fatalf("View must not contain \\r; Bubble Tea blank-screens on CR")
	}
	lines := strings.Split(view, "\n")
	titleIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "Select context") {
			titleIdx = i
			break
		}
	}
	if titleIdx < 0 {
		t.Fatalf("expected Select context in centered overlay, got:\n%s", view)
	}
	// With height 40, a ~12-row modal should start near the vertical middle (~14),
	// not stacked under the workspace (~row 25+) or stuck at the bottom.
	if titleIdx < 8 || titleIdx > 22 {
		t.Fatalf("expected modal title near vertical center, got row %d of %d:\n%s", titleIdx, len(lines), view)
	}
	if !strings.Contains(view, "Catalog") {
		t.Fatalf("expected workspace behind overlay to remain visible, got:\n%s", view)
	}
}
