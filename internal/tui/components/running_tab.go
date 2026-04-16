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
	if len(entries) == 0 {
		return "Running tab (empty)"
	}
	var b strings.Builder
	b.WriteString("Running:\n")
	for _, entry := range entries {
		line := fmt.Sprintf("  • %s  %d→%d  [%s]", entry.Label, entry.LocalPort, entry.RemotePort, entry.Status)
		if entry.Err != "" {
			line += "  err=" + entry.Err
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}
