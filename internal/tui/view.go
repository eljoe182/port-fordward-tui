package tui

import (
	"fmt"
	"strings"

	"port-forward-tui/internal/tui/components"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var (
	workspaceStyle   = lipgloss.NewStyle().Padding(0, 1)
	panelStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	activeTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	inactiveTabStyle = lipgloss.NewStyle().Faint(true)
	panelHintStyle   = lipgloss.NewStyle().Faint(true)
)

const (
	catalogPaneWidth = 76
	sidePaneWidth    = 52
	fixedPaneHeight  = 18
)

func (m Model) View() string {
	header := components.Header(components.HeaderData{
		Query:  m.query,
		Filter: string(m.filterMode),
		Sort:   string(m.sortMode),
		Err:    m.errMsg,
	})
	footer := components.Footer(m.contextName, m.namespace)
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
	var panelHints string
	switch m.activeTab {
	case TabRunning:
		runningEntries := make([]components.RunningEntry, 0, len(m.running))
		for _, entry := range m.running {
			runningEntries = append(runningEntries, components.RunningEntry{
				Context:    entry.Context,
				Namespace:  entry.Namespace,
				Label:      entry.Label,
				LocalPort:  entry.LocalPort,
				RemotePort: entry.RemotePort,
				Status:     string(entry.Status),
				Err:        entry.Err,
			})
		}
		panelBody = components.RunningTabWindow(runningEntries, m.runningCursor, bodyHeight-4)
		panelHints = components.PanelHintsRunning
	default:
		selectedEntries := make([]components.SelectedEntry, 0, len(m.selected))
		for _, entry := range m.selected {
			selectedEntries = append(selectedEntries, components.SelectedEntry{
				Context:    entry.Context,
				Namespace:  entry.Namespace,
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
		}, bodyHeight-4)
		panelHints = components.PanelHintsSelected
	}
	if panelHints != "" {
		panelBody = strings.TrimRight(panelBody, "\n") + "\n\n" + panelHintStyle.Render(panelHints)
	}
	rightPane := renderPane(renderTabs(m.activeTab), panelBody, sidePaneWidth, bodyHeight)
	workspace := workspaceStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, catalog, rightPane))
	base := header + "\n" + workspace + "\n" + footer

	modal := m.renderModal()
	if modal == "" {
		return base
	}

	width := m.width
	height := m.height
	baseWidth := lipgloss.Width(base)
	baseHeight := lipgloss.Height(base)
	if width < baseWidth {
		width = baseWidth
	}
	if height < baseHeight {
		height = baseHeight
	}
	if width <= 0 {
		width = baseWidth
	}
	if height <= 0 {
		height = baseHeight
	}
	return placeOverlay(width, height, base, modal)
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

// placeOverlay composites overlay centered on top of base. Output uses \n only
// (Bubble Tea breaks if View contains \r from sources like cellbuf.Render).
func placeOverlay(width, height int, base, overlay string) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	base = normalizeNewlines(base)
	overlay = normalizeNewlines(overlay)

	baseLines := strings.Split(strings.TrimRight(base, "\n"), "\n")
	for len(baseLines) < height {
		baseLines = append(baseLines, "")
	}
	if len(baseLines) > height {
		baseLines = baseLines[:height]
	}
	for i := range baseLines {
		baseLines[i] = padVisual(baseLines[i], width)
	}

	overlayLines := strings.Split(strings.TrimRight(overlay, "\n"), "\n")
	ow := 0
	for _, line := range overlayLines {
		if w := ansi.StringWidth(line); w > ow {
			ow = w
		}
	}
	oh := len(overlayLines)
	if ow > width {
		ow = width
	}
	if oh > height {
		oh = height
		overlayLines = overlayLines[:oh]
	}
	x := (width - ow) / 2
	y := (height - oh) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	for i, line := range overlayLines {
		row := y + i
		if row < 0 || row >= len(baseLines) {
			continue
		}
		mid := padVisual(line, ow)
		if ansi.StringWidth(mid) > ow {
			mid = ansi.Truncate(mid, ow, "")
		}
		left := ansi.Cut(baseLines[row], 0, x)
		right := ansi.Cut(baseLines[row], x+ow, width+1)
		baseLines[row] = left + mid + right
	}
	return strings.Join(baseLines, "\n")
}

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "")
}

func padVisual(line string, width int) string {
	w := ansi.StringWidth(line)
	if w > width {
		return ansi.Truncate(line, width, "")
	}
	if w < width {
		return line + strings.Repeat(" ", width-w)
	}
	return line
}
