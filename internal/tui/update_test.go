package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTabKeySwitchesFromSelectedToRunning(t *testing.T) {
	m := NewModel(Dependencies{})
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated := next.(Model)

	if updated.activeTab != TabRunning {
		t.Fatalf("expected running tab, got %q", updated.activeTab)
	}
}
