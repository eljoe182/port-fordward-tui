package components

import "github.com/charmbracelet/lipgloss"

var footerStyle = lipgloss.NewStyle().Faint(true).Padding(0, 1)

func Footer(activeTab string) string {
	var hints string
	switch activeTab {
	case "running":
		hints = "J/K nav • x stop • R retry • / search • t filter • o sort • c ctx • n ns • r refresh • tab switch • q quit"
	default:
		hints = "↑/↓ nav • enter select • f favorite • / search • t filter • o sort • J/K tab-cursor • e edit port • s start • c ctx • n ns • r refresh • tab switch • q quit"
	}
	return footerStyle.Render(hints)
}
