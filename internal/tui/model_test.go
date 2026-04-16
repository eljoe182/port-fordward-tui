package tui

import "testing"

func TestNewModelStartsWithSelectedTabAndEmptyState(t *testing.T) {
	model := NewModel(Dependencies{})

	if model.activeTab != TabSelected {
		t.Fatalf("expected default tab %q, got %q", TabSelected, model.activeTab)
	}

	if len(model.catalog) != 0 {
		t.Fatalf("expected empty catalog on startup")
	}
}
