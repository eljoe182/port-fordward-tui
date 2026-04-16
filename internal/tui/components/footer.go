package components

import "github.com/charmbracelet/lipgloss"

var footerStyle = lipgloss.NewStyle().Faint(true).Padding(0, 1)

func Footer(activeTab string) string {
	var hints string
	switch activeTab {
	case "running":
		hints = "J/K nav • tab switch • x stop • q quit"
	default:
		hints = "↑/↓ nav • enter select • J/K tab-cursor • e edit port • s start • tab switch • q quit"
	}
	return footerStyle.Render(hints)
}
