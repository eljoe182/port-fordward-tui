package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var footerStyle = lipgloss.NewStyle().Faint(true).Padding(0, 1)

func Footer(contextName, namespace string) string {
	status := fmt.Sprintf("ctx=%s  ns=%s", orDash(contextName), orDash(namespace))
	hints := "↑/↓ nav • enter select • f favorite • / search • t filter • o sort • c ctx • n ns • tab switch • ? extra"
	return footerStyle.Render(status + "\n" + hints)
}

// PanelHintsSelected are local shortcuts shown at the bottom of the Selected panel.
const PanelHintsSelected = "j/k tab-cursor • tab switch • e edit port • s start • x remove/stop • r refresh"

// PanelHintsRunning are local shortcuts shown at the bottom of the Running panel.
const PanelHintsRunning = "j/k tab-cursor • tab switch • x stop • R retry • r refresh"
