package tui

import "testing"

type tuiLocalPortChecker struct {
	unavailable map[int]bool
}

func (f tuiLocalPortChecker) Available(port int) bool {
	return !f.unavailable[port]
}

func TestSelectCurrentItemSkipsReservedAndHostUnavailablePorts(t *testing.T) {
	m := NewModel(Dependencies{LocalPorts: tuiLocalPortChecker{unavailable: map[int]bool{3002: true}}})
	m.selected = []SelectedItem{{TargetID: "service:default:one", LocalPort: 3000}}
	m.running = []RunningItem{{TargetID: "service:default:two", LocalPort: 3001, Status: StatusStarting}}
	m.catalog = []CatalogItem{{
		Context: "prod", Namespace: "default", Type: "service", ID: "service:default:three", Label: "three", PreferredLocalPort: 3000,
	}}

	m.selectCurrentItem()

	if len(m.selected) != 2 || m.selected[1].LocalPort != 3003 {
		t.Fatalf("expected fallback port 3003, got %#v", m.selected)
	}
}

func TestSelectCurrentItemReportsExhaustedPortRange(t *testing.T) {
	m := NewModel(Dependencies{LocalPorts: tuiLocalPortChecker{}})
	m.selected = []SelectedItem{{TargetID: "service:default:one", LocalPort: 65535}}
	m.catalog = []CatalogItem{{
		Context: "prod", Namespace: "default", Type: "service", ID: "service:default:two", Label: "two", PreferredLocalPort: 65535,
	}}

	m.selectCurrentItem()

	if len(m.selected) != 1 {
		t.Fatalf("expected exhausted target not to be selected, got %#v", m.selected)
	}
	if m.errMsg == "" {
		t.Fatalf("expected actionable port allocation error")
	}
}
