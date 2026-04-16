package tui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"cco-port-forward-tui/internal/domain"
	"cco-port-forward-tui/internal/ports"
)

type forwardStartedMsg struct {
	TargetID  string
	SessionID string
}

type forwardFailedMsg struct {
	TargetID string
	Err      string
}

type forwardStoppedMsg struct {
	TargetID string
}

func startForwardsCmds(ctx context.Context, runner ports.ForwardRunner, selected []SelectedItem, contextName, namespace string) []tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(selected))
	for _, item := range selected {
		req := domain.ForwardRequest{
			TargetID:   item.TargetID,
			Label:      item.Label,
			LocalPort:  item.LocalPort,
			RemotePort: item.RemotePort,
			Context:    contextName,
			Namespace:  namespace,
			Type:       targetTypeFromID(item.TargetID),
		}
		cmds = append(cmds, startForwardCmd(ctx, runner, req))
	}
	return cmds
}

func startForwardCmd(ctx context.Context, runner ports.ForwardRunner, req domain.ForwardRequest) tea.Cmd {
	return func() tea.Msg {
		sessionID, err := runner.Start(ctx, req)
		if err != nil {
			return forwardFailedMsg{TargetID: req.TargetID, Err: err.Error()}
		}
		return forwardStartedMsg{TargetID: req.TargetID, SessionID: sessionID}
	}
}

func stopForwardCmd(ctx context.Context, runner ports.ForwardRunner, targetID, sessionID string) tea.Cmd {
	return func() tea.Msg {
		if err := runner.Stop(ctx, sessionID); err != nil {
			return forwardFailedMsg{TargetID: targetID, Err: err.Error()}
		}
		return forwardStoppedMsg{TargetID: targetID}
	}
}

func targetTypeFromID(id string) domain.TargetType {
	prefix := strings.SplitN(id, ":", 2)[0]
	switch prefix {
	case string(domain.TargetTypePod):
		return domain.TargetTypePod
	case string(domain.TargetTypeService):
		return domain.TargetTypeService
	}
	return domain.TargetTypeService
}
