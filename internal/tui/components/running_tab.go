package components

import (
	"fmt"
	"strings"
)

type RunningEntry struct {
	Label      string
	LocalPort  int
	RemotePort int
	Status     string
	Err        string
}

func RunningTab(entries ...RunningEntry) string {
	return RunningTabWindow(entries, 0, len(entries))
}

func RunningTabWindow(entries []RunningEntry, cursor, maxRows int) string {
	if len(entries) == 0 {
		return "Running tab (empty)"
	}
	start, end := visibleWindow(len(entries), cursor, maxRows)
	var b strings.Builder
	if start > 0 {
		b.WriteString(fmt.Sprintf("  ↑ %d more\n", start))
	}
	for i := start; i < end; i++ {
		entry := entries[i]
		marker := "  "
		if i == cursor {
			marker = "▶ "
		}
		line := fmt.Sprintf("  • %s  %d→%d  [%s]", entry.Label, entry.LocalPort, entry.RemotePort, entry.Status)
		line = marker + strings.TrimPrefix(line, "  ")
		if entry.Err != "" {
			line += "  err=" + entry.Err
			if entry.Status == "failed" {
				line += "  (press R to retry)"
			}
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	if end < len(entries) {
		b.WriteString(fmt.Sprintf("  ↓ %d more\n", len(entries)-end))
	}
	return b.String()
}
