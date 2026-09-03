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
		return "Name | Type | NS | Port\n  (no targets)"
	}
	// Reserve one row for the column header when clipping.
	contentRows := maxRows
	if contentRows > 1 {
		contentRows--
	}
	start, end := visibleWindow(len(items), cursor, contentRows)
	var b strings.Builder
	b.WriteString("Name | Type | NS | Port\n")
	if start > 0 {
		b.WriteString(fmt.Sprintf("  ↑ %d more\n", start))
	}
	for i := start; i < end; i++ {
		item := items[i]
		marker := "  "
		if i == cursor {
			marker = "> "
		}
		ns := item.Namespace
		if ns == "" {
			ns = "-"
		}
		port := formatCatalogPort(item)
		typeLabel := item.Type
		if item.Favorite {
			typeLabel += " ★"
		}
		if !item.Available {
			typeLabel += " unavailable"
		}
		b.WriteString(fmt.Sprintf("%s%s | %s | %s | %s\n", marker, item.Label, typeLabel, ns, port))
	}
	if end < len(items) {
		b.WriteString(fmt.Sprintf("  ↓ %d more\n", len(items)-end))
	}
	return b.String()
}

func formatCatalogPort(item Item) string {
	if item.PreferredLocalPort != 0 {
		return fmt.Sprintf("%d→%d", item.PreferredLocalPort, item.RemotePort)
	}
	if item.RemotePort != 0 {
		return fmt.Sprintf("%d", item.RemotePort)
	}
	return "-"
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
		if start < 0 {
			start = 0
		}
	}
	return start, end
}
