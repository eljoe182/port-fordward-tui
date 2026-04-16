package tui

import (
	"fmt"

	"port-forward-tui/internal/tui/components"

	"github.com/charmbracelet/lipgloss"
)

var (
	workspaceStyle   = lipgloss.NewStyle().Padding(0, 1)
	panelStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	activeTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	inactiveTabStyle = lipgloss.NewStyle().Faint(true)
	modalStyle       = lipgloss.NewStyle().Border(lipgloss.ThickBorder()).Padding(1, 2).Width(54).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("236"))
	modalActiveStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
)

const (
	catalogPaneWidth = 76
	sidePaneWidth    = 52
	fixedPaneHeight  = 18
)

func (m Model) View() string {
	header := components.Header(components.HeaderData{
		ActiveTab:   string(m.activeTab),
		Context:     m.contextName,
		Namespace:   m.namespace,
		Query:       m.query,
		Filter:      string(m.filterMode),
		Sort:        string(m.sortMode),
		Searching:   m.modalKind == ModalSearch,
		QueryBuffer: m.modalInput,
		Err:         m.errMsg,
	})
	footer := components.Footer(string(m.activeTab))
	bodyHeight := fixedPaneHeight

	catalogItems := make([]components.Item, 0, len(m.catalog))
	for _, item := range m.catalog {
		catalogItems = append(catalogItems, components.Item{
			Type:               item.Type,
			Label:              item.Label,
			Namespace:          item.Namespace,
			PreferredLocalPort: item.PreferredLocalPort,
			RemotePort:         item.RemotePort,
			Favorite:           item.Favorite,
			Available:          item.Available,
		})
	}
	catalog := renderPane("Catalog", components.CatalogWindow(catalogItems, m.cursor, bodyHeight-2), catalogPaneWidth, bodyHeight)

	var panelBody string
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
		panelBody = components.RunningTabWindow(runningEntries, m.runningCursor, bodyHeight-2)
	default:
		selectedEntries := make([]components.SelectedEntry, 0, len(m.selected))
		for _, entry := range m.selected {
			selectedEntries = append(selectedEntries, components.SelectedEntry{
				Label:      entry.Label,
				LocalPort:  entry.LocalPort,
				RemotePort: entry.RemotePort,
			})
		}
		panelBody = components.SelectedTabWindow(components.SelectedTabData{
			Entries:     selectedEntries,
			Cursor:      m.selectedCursor,
			EditingPort: m.editingPort,
			PortBuffer:  m.portBuffer,
		}, bodyHeight-2)
	}
	rightPane := renderPane(renderTabs(m.activeTab), panelBody, sidePaneWidth, bodyHeight)
	workspace := workspaceStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, catalog, rightPane))
	if modal := m.renderModal(); modal != "" {
		return header + "\n" + workspace + "\n\n" + modalStyle.Render(modal) + "\n" + footer
	}
	return header + "\n" + workspace + "\n" + footer
}

func renderPane(title, body string, width, height int) string {
	header := lipgloss.NewStyle().Bold(true).Width(width - 4).Render(title)
	contentHeight := height - lipgloss.Height(header) - 2
	if contentHeight < 3 {
		contentHeight = 3
	}
	content := lipgloss.NewStyle().Width(width - 4).Height(contentHeight).Render(body)
	return panelStyle.Width(width).Render(fmt.Sprintf("%s\n\n%s", header, content))
}

func renderTabs(active Tab) string {
	selected := inactiveTabStyle.Render("Selected")
	running := inactiveTabStyle.Render("Running")
	if active == TabSelected {
		selected = activeTabStyle.Render("Selected")
	} else {
		running = activeTabStyle.Render("Running")
	}
	return fmt.Sprintf("Panel  [%s] [%s]", selected, running)
}

func renderSelectorModal(title string, options []modalOption, cursor int) string {
	var out string
	out += lipgloss.NewStyle().Bold(true).Render(title) + "\n\n"
	for i, option := range options {
		prefix := "  "
		label := option.Label
		if i == cursor {
			prefix = "▶ "
			label = modalActiveStyle.Render(label)
		}
		out += prefix + label + "\n"
	}
	out += "\nEnter select • Esc cancel • ↑/↓ or j/k navigate"
	return out
}
