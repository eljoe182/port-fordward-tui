package tui

import (
	"cco-port-forward-tui/internal/tui/components"
)

func (m Model) View() string {
	header := components.Header(string(m.activeTab))
	catalog := components.Catalog(m.catalog)
	var body string
	switch m.activeTab {
	case TabRunning:
		body = components.RunningTab()
	default:
		body = components.SelectedTab()
	}
	return header + "\n" + catalog + "\n" + body
}
