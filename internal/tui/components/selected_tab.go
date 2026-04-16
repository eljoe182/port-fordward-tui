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
	return SelectedTabWindow(data, len(data.Entries))
}

func SelectedTabWindow(data SelectedTabData, maxRows int) string {
	if len(data.Entries) == 0 {
		return "Selected tab (empty)"
	}
	start, end := visibleWindow(len(data.Entries), data.Cursor, maxRows)
	var b strings.Builder
	if start > 0 {
		b.WriteString(fmt.Sprintf("  ↑ %d more\n", start))
	}
	for i := start; i < end; i++ {
		entry := data.Entries[i]
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
	if end < len(data.Entries) {
		b.WriteString(fmt.Sprintf("  ↓ %d more\n", len(data.Entries)-end))
	}
	if data.EditingPort {
		b.WriteString("\n(editing local port — digits to type, Enter commit, Esc cancel)\n")
	}
	return b.String()
}
