package tui

import (
	"context"
	"strings"

	appruntime "cco-port-forward-tui/internal/app/runtime"
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

type forwardBatchMsg struct {
	started []forwardStartedMsg
	failed  []forwardFailedMsg
}

func retryForwardCmd(ctx context.Context, svc appruntime.Service, item RunningItem, active []domain.ForwardSession) tea.Cmd {
	return func() tea.Msg {
		result, err := svc.StartOne(ctx, runningToRequest(item), active)
		if err != nil {
			return forwardFailedMsg{TargetID: item.TargetID, Err: err.Error()}
		}
		if result.Err != nil {
			return forwardFailedMsg{TargetID: item.TargetID, Err: result.Err.Error()}
		}
		return forwardStartedMsg{TargetID: item.TargetID, SessionID: result.SessionID}
	}
}

func startForwardsCmd(ctx context.Context, svc appruntime.Service, selected []SelectedItem, contextName, namespace string, active []domain.ForwardSession) tea.Cmd {
	requests := make([]domain.ForwardRequest, 0, len(selected))
	for _, item := range selected {
		requests = append(requests, selectedToRequest(item, contextName, namespace))
	}
	return func() tea.Msg {
		results, err := svc.StartMany(ctx, requests, active)
		if err != nil {
			return catalogErrorMsg{err: err}
		}
		msg := forwardBatchMsg{}
		for _, result := range results {
			if result.Err != nil {
				msg.failed = append(msg.failed, forwardFailedMsg{TargetID: result.Request.TargetID, Err: result.Err.Error()})
				continue
			}
			msg.started = append(msg.started, forwardStartedMsg{TargetID: result.Request.TargetID, SessionID: result.SessionID})
		}
		return msg
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

func selectedToRequest(item SelectedItem, contextName, namespace string) domain.ForwardRequest {
	return domain.ForwardRequest{
		TargetID:   item.TargetID,
		Label:      item.Label,
		LocalPort:  item.LocalPort,
		RemotePort: item.RemotePort,
		Context:    contextName,
		Namespace:  namespace,
		Type:       targetTypeFromID(item.TargetID),
	}
}

func runningToRequest(item RunningItem) domain.ForwardRequest {
	return domain.ForwardRequest{
		TargetID:   item.TargetID,
		Label:      item.Label,
		LocalPort:  item.LocalPort,
		RemotePort: item.RemotePort,
		Context:    item.Context,
		Namespace:  item.Namespace,
		Type:       domain.TargetType(item.Type),
	}
}

func activeForwardSessions(items []RunningItem) []domain.ForwardSession {
	sessions := make([]domain.ForwardSession, 0, len(items))
	for _, item := range items {
		sessions = append(sessions, domain.ForwardSession{
			TargetID:   item.TargetID,
			Label:      item.Label,
			LocalPort:  item.LocalPort,
			RemotePort: item.RemotePort,
			Status:     domain.ForwardStatus(item.Status),
			Err:        item.Err,
		})
	}
	return sessions
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
