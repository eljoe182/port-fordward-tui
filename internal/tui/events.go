package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"port-forward-tui/internal/domain"
	"port-forward-tui/internal/ports"
)

type forwardEventMsg struct {
	event domain.ForwardEvent
}

func listenForwardEventsCmd(runner ports.ForwardRunner) tea.Cmd {
	if runner == nil {
		return nil
	}
	events := runner.Events()
	if events == nil {
		return nil
	}
	return func() tea.Msg {
		event, ok := <-events
		if !ok {
			return nil
		}
		return forwardEventMsg{event: event}
	}
}
