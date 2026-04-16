package components

import "strings"

func Catalog(items []string) string {
	if len(items) == 0 {
		return "  (no targets)"
	}
	var b strings.Builder
	b.WriteString("Catalog:\n")
	for _, item := range items {
		b.WriteString("  • ")
		b.WriteString(item)
		b.WriteString("\n")
	}
	return b.String()
}
