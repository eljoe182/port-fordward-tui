package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSelectedTabCursorMovesWithShiftJAndK(t *testing.T) {
	m := NewModel(Dependencies{})
	m.activeTab = TabSelected
	m.selected = []SelectedItem{
		{TargetID: "service:a", Label: "a", LocalPort: 3001, RemotePort: 3000},
		{TargetID: "service:b", Label: "b", LocalPort: 3002, RemotePort: 3000},
		{TargetID: "service:c", Label: "c", LocalPort: 3003, RemotePort: 3000},
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("J")})
	m = next.(Model)
	if m.selectedCursor != 1 {
		t.Fatalf("expected selectedCursor 1 after J, got %d", m.selectedCursor)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("K")})
	m = next.(Model)
	if m.selectedCursor != 0 {
		t.Fatalf("expected selectedCursor 0 after K, got %d", m.selectedCursor)
	}
}

func TestEKeyOnSelectedTabEntersPortEditMode(t *testing.T) {
	m := NewModel(Dependencies{})
	m.activeTab = TabSelected
	m.selected = []SelectedItem{
		{TargetID: "service:a", Label: "a", LocalPort: 3001, RemotePort: 3000},
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = next.(Model)

	if !m.editingPort {
		t.Fatalf("expected editingPort true")
	}
	if m.portBuffer != "3001" {
		t.Fatalf("expected buffer seeded with current port, got %q", m.portBuffer)
	}
}

func TestEnterInPortEditModeCommitsValidPort(t *testing.T) {
	m := NewModel(Dependencies{})
	m.activeTab = TabSelected
	m.selected = []SelectedItem{
		{TargetID: "service:a", Label: "a", LocalPort: 3001, RemotePort: 3000},
	}
	m.editingPort = true
	m.portBuffer = "8080"

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)

	if m.editingPort {
		t.Fatalf("expected edit mode to exit after commit")
	}
	if m.selected[0].LocalPort != 8080 {
		t.Fatalf("expected LocalPort=8080, got %d", m.selected[0].LocalPort)
	}
}

func TestEscInPortEditModeCancelsWithoutCommitting(t *testing.T) {
	m := NewModel(Dependencies{})
	m.activeTab = TabSelected
	m.selected = []SelectedItem{
		{TargetID: "service:a", Label: "a", LocalPort: 3001, RemotePort: 3000},
	}
	m.editingPort = true
	m.portBuffer = "8080"

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)

	if m.editingPort {
		t.Fatalf("expected edit mode to exit after esc")
	}
	if m.selected[0].LocalPort != 3001 {
		t.Fatalf("expected LocalPort unchanged=3001, got %d", m.selected[0].LocalPort)
	}
}

func TestDigitKeysAppendToPortBufferWhileEditing(t *testing.T) {
	m := NewModel(Dependencies{})
	m.activeTab = TabSelected
	m.selected = []SelectedItem{
		{TargetID: "service:a", Label: "a", LocalPort: 3001, RemotePort: 3000},
	}
	m.editingPort = true
	m.portBuffer = ""

	for _, r := range []rune("9090") {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(Model)
	}

	if m.portBuffer != "9090" {
		t.Fatalf("expected buffer 9090, got %q", m.portBuffer)
	}
}

func TestBackspaceRemovesLastDigitFromPortBuffer(t *testing.T) {
	m := NewModel(Dependencies{})
	m.activeTab = TabSelected
	m.selected = []SelectedItem{
		{TargetID: "service:a", Label: "a", LocalPort: 3001, RemotePort: 3000},
	}
	m.editingPort = true
	m.portBuffer = "3001"

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = next.(Model)

	if m.portBuffer != "300" {
		t.Fatalf("expected buffer 300, got %q", m.portBuffer)
	}
}

func TestEnterWithInvalidPortKeepsEditModeAndSetsError(t *testing.T) {
	m := NewModel(Dependencies{})
	m.activeTab = TabSelected
	m.selected = []SelectedItem{
		{TargetID: "service:a", Label: "a", LocalPort: 3001, RemotePort: 3000},
	}
	m.editingPort = true
	m.portBuffer = "99999"

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)

	if !m.editingPort {
		t.Fatalf("expected to stay in edit mode on invalid port")
	}
	if m.selected[0].LocalPort != 3001 {
		t.Fatalf("expected LocalPort unchanged on invalid commit, got %d", m.selected[0].LocalPort)
	}
	if m.errMsg == "" {
		t.Fatalf("expected errMsg populated on invalid port")
	}
}
