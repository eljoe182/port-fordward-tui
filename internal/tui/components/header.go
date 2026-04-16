package components

import "github.com/charmbracelet/lipgloss"

var headerStyle = lipgloss.NewStyle().Bold(true).Padding(0, 1)

func Header(activeTab string) string {
	return headerStyle.Render("portfwd-tui • tab: " + activeTab)
}
