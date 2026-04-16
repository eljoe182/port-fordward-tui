package tui

import (
	"cco-port-forward-tui/internal/tui/components"
)

func (m Model) View() string {
	header := components.Header(string(m.activeTab))

	catalogItems := make([]components.Item, 0, len(m.catalog))
	for _, item := range m.catalog {
		catalogItems = append(catalogItems, components.Item{
			Label:              item.Label,
			PreferredLocalPort: item.PreferredLocalPort,
			RemotePort:         item.RemotePort,
		})
	}
	catalog := components.Catalog(catalogItems, m.cursor)

	var body string
	switch m.activeTab {
	case TabRunning:
		runningEntries := make([]components.RunningEntry, 0, len(m.running))
		for _, entry := range m.running {
			runningEntries = append(runningEntries, components.RunningEntry{
				Label:      entry.Label,
				LocalPort:  entry.LocalPort,
				RemotePort: entry.RemotePort,
				Status:     string(entry.Status),
				Err:        entry.Err,
			})
		}
		body = components.RunningTab(runningEntries...)
	default:
		selectedEntries := make([]components.SelectedEntry, 0, len(m.selected))
		for _, entry := range m.selected {
			selectedEntries = append(selectedEntries, components.SelectedEntry{
				Label:      entry.Label,
				LocalPort:  entry.LocalPort,
				RemotePort: entry.RemotePort,
			})
		}
		body = components.SelectedTab(selectedEntries)
	}
	return header + "\n" + catalog + "\n" + body
}
