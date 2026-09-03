package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	metaStyle   = lipgloss.NewStyle().Faint(true).Padding(0, 1)
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Padding(0, 1)
)

type HeaderData struct {
	Query  string
	Filter string
	Sort   string
	Err    string
}

func Header(data HeaderData) string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("portfwd-tui"))
	meta := fmt.Sprintf("filter=%s  sort=%s  query=%s",
		orDash(data.Filter), orDash(data.Sort), orDash(data.Query))
	b.WriteString(metaStyle.Render(meta))
	if data.Err != "" {
		b.WriteString("\n")
		b.WriteString(errStyle.Render("error: " + data.Err))
	}
	return b.String()
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
