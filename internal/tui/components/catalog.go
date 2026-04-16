package components

import (
	"fmt"
	"strings"
)

type Item struct {
	Type               string
	Label              string
	Namespace          string
	PreferredLocalPort int
	RemotePort         int
	Favorite           bool
	Available          bool
}

func Catalog(items []Item, cursor int) string {
	return CatalogWindow(items, cursor, len(items))
}

func CatalogWindow(items []Item, cursor, maxRows int) string {
	if len(items) == 0 {
		return "  (no targets)"
	}
	start, end := visibleWindow(len(items), cursor, maxRows)
	var b strings.Builder
	if start > 0 {
		b.WriteString(fmt.Sprintf("  ↑ %d more\n", start))
	}
	for i := start; i < end; i++ {
		item := items[i]
		marker := "  "
		if i == cursor {
			marker = "> "
		}
		label := item.Label
		meta := fmt.Sprintf("[%s]", item.Type)
		if item.Namespace != "" {
			meta += " ns=" + item.Namespace
		}
		if item.PreferredLocalPort != 0 {
			meta += fmt.Sprintf("  %d→%d", item.PreferredLocalPort, item.RemotePort)
		}
		if item.Favorite {
			meta += "  ★"
		}
		if !item.Available {
			meta += "  unavailable"
		}
		b.WriteString(marker)
		b.WriteString(label)
		b.WriteString(" ")
		b.WriteString(meta)
		b.WriteString("\n")
	}
	if end < len(items) {
		b.WriteString(fmt.Sprintf("  ↓ %d more\n", len(items)-end))
	}
	return b.String()
}

func visibleWindow(total, cursor, maxRows int) (int, int) {
	if maxRows <= 0 || total <= maxRows {
		return 0, total
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= total {
		cursor = total - 1
	}
	half := maxRows / 2
	start := cursor - half
	if start < 0 {
		start = 0
	}
	end := start + maxRows
	if end > total {
		end = total
		start = end - maxRows
	}
	if start < 0 {
		start = 0
	}
	return start, end
}
