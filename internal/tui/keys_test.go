package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newModelWithCatalog(items ...CatalogItem) Model {
	m := NewModel(Dependencies{})
	m.catalog = items
	return m
}

func pressKey(t *testing.T, m Model, key string) Model {
	t.Helper()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	next, _ := m.Update(msg)
	return next.(Model)
}

func pressSpecial(t *testing.T, m Model, keyType tea.KeyType) Model {
	t.Helper()
	next, _ := m.Update(tea.KeyMsg{Type: keyType})
	return next.(Model)
}

func TestCursorMovesDownAndUpWithinCatalog(t *testing.T) {
	m := newModelWithCatalog(
		CatalogItem{ID: "a", Label: "a"},
		CatalogItem{ID: "b", Label: "b"},
		CatalogItem{ID: "c", Label: "c"},
	)

	m = pressKey(t, m, "j")
	if m.cursor != 1 {
		t.Fatalf("expected cursor 1 after j, got %d", m.cursor)
	}
	m = pressSpecial(t, m, tea.KeyDown)
	if m.cursor != 2 {
		t.Fatalf("expected cursor 2 after down, got %d", m.cursor)
	}
	m = pressKey(t, m, "j")
	if m.cursor != 2 {
		t.Fatalf("cursor must clamp at bottom, got %d", m.cursor)
	}
	m = pressKey(t, m, "k")
	if m.cursor != 1 {
		t.Fatalf("expected cursor 1 after k, got %d", m.cursor)
	}
	m = pressSpecial(t, m, tea.KeyUp)
	m = pressSpecial(t, m, tea.KeyUp)
	if m.cursor != 0 {
		t.Fatalf("cursor must clamp at top, got %d", m.cursor)
	}
}

func TestEnterAddsCurrentCatalogItemToSelected(t *testing.T) {
	m := newModelWithCatalog(
		CatalogItem{ID: "service:admin", Label: "admin", RemotePort: 3000, PreferredLocalPort: 3001},
	)

	m = pressSpecial(t, m, tea.KeyEnter)

	if len(m.selected) != 1 {
		t.Fatalf("expected 1 selected, got %d", len(m.selected))
	}
	if m.selected[0].LocalPort != 3001 {
		t.Fatalf("expected preferred local port 3001, got %d", m.selected[0].LocalPort)
	}
}

func TestEnterDoesNotDuplicateAlreadySelectedItem(t *testing.T) {
	m := newModelWithCatalog(
		CatalogItem{ID: "service:admin", Label: "admin", RemotePort: 3000, PreferredLocalPort: 3001},
	)

	m = pressSpecial(t, m, tea.KeyEnter)
	m = pressSpecial(t, m, tea.KeyEnter)

	if len(m.selected) != 1 {
		t.Fatalf("expected selected to stay at 1, got %d", len(m.selected))
	}
}
