package components

import (
	"fmt"
	"strings"
)

type SelectedEntry struct {
	Label      string
	LocalPort  int
	RemotePort int
}

func SelectedTab(entries []SelectedEntry) string {
	if len(entries) == 0 {
		return "Selected tab (empty)"
	}
	var b strings.Builder
	b.WriteString("Selected:\n")
	for _, entry := range entries {
		b.WriteString(fmt.Sprintf("  • %s  %d→%d\n", entry.Label, entry.LocalPort, entry.RemotePort))
	}
	return b.String()
}
