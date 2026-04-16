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

type SelectedTabData struct {
	Entries     []SelectedEntry
	Cursor      int
	EditingPort bool
	PortBuffer  string
}

func SelectedTab(data SelectedTabData) string {
	if len(data.Entries) == 0 {
		return "Selected tab (empty)"
	}
	var b strings.Builder
	b.WriteString("Selected:\n")
	for i, entry := range data.Entries {
		marker := "  "
		if i == data.Cursor {
			marker = "▶ "
		}
		localPort := fmt.Sprintf("%d", entry.LocalPort)
		if i == data.Cursor && data.EditingPort {
			localPort = "[" + data.PortBuffer + "_]"
		}
		b.WriteString(fmt.Sprintf("%s%s  %s→%d\n", marker, entry.Label, localPort, entry.RemotePort))
	}
	if data.EditingPort {
		b.WriteString("\n(editing local port — digits to type, Enter commit, Esc cancel)\n")
	}
	return b.String()
}
