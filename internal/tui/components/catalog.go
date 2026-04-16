package components

import (
	"fmt"
	"strings"
)

type Item struct {
	Label              string
	PreferredLocalPort int
	RemotePort         int
}

func Catalog(items []Item, cursor int) string {
	if len(items) == 0 {
		return "  (no targets)"
	}
	var b strings.Builder
	b.WriteString("Catalog:\n")
	for i, item := range items {
		marker := "  "
		if i == cursor {
			marker = "> "
		}
		label := item.Label
		if item.PreferredLocalPort != 0 {
			label = fmt.Sprintf("%s  %d→%d", label, item.PreferredLocalPort, item.RemotePort)
		}
		b.WriteString(marker)
		b.WriteString(label)
		b.WriteString("\n")
	}
	return b.String()
}
