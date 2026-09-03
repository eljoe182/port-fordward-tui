package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	modalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(1, 2).
			MaxWidth(72).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("236"))
	modalTitleStyle  = lipgloss.NewStyle().Bold(true)
	modalActiveStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	modalHintStyle   = lipgloss.NewStyle().Faint(true)
)

// ModalOption is a labeled choice in a list picker modal.
type ModalOption struct {
	Label string
}

// ListModal renders a bordered selector with cursor and navigation hints.
func ListModal(title string, options []ModalOption, cursor int) string {
	var b strings.Builder
	b.WriteString(modalTitleStyle.Render(title))
	b.WriteString("\n\n")
	for i, option := range options {
		prefix := "  "
		label := option.Label
		if i == cursor {
			prefix = "▶ "
			label = modalActiveStyle.Render(label)
		}
		b.WriteString(prefix)
		b.WriteString(label)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(modalHintStyle.Render("Enter select • Esc cancel • ↑/↓ or j/k navigate"))
	return modalBoxStyle.Render(b.String())
}

// TextModal renders a bordered text-input modal (search).
func TextModal(title, input string) string {
	body := fmt.Sprintf("%s\n\nquery: [%s_]\n\n%s",
		modalTitleStyle.Render(title),
		input,
		modalHintStyle.Render("Enter apply • Esc cancel"),
	)
	return modalBoxStyle.Render(body)
}

// HelpModal renders a bordered static help panel (Extra).
func HelpModal(title string, lines []string) string {
	var b strings.Builder
	b.WriteString(modalTitleStyle.Render(title))
	b.WriteString("\n\n")
	for _, line := range lines {
		b.WriteString("  ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(modalHintStyle.Render("Esc close"))
	return modalBoxStyle.Render(b.String())
}
