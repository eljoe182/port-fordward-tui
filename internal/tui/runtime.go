package tui

import (
	"context"

	appruntime "port-forward-tui/internal/app/runtime"

	tea "github.com/charmbracelet/bubbletea"

	"port-forward-tui/internal/domain"
	"port-forward-tui/internal/ports"
)

type forwardStartedMsg struct {
	Ref       forwardRef
	SessionID string
}

type forwardFailedMsg struct {
	Ref forwardRef
	Err string
}

type forwardStoppedMsg struct {
	SessionID string
}

type forwardStopFailedMsg struct {
	SessionID string
	Err       string
}

type forwardBatchMsg struct {
	started []forwardStartedMsg
	failed  []forwardFailedMsg
}

func retryForwardCmd(ctx context.Context, svc appruntime.Service, item RunningItem, active []domain.ForwardSession) tea.Cmd {
	return func() tea.Msg {
		result, err := svc.StartOne(ctx, runningToRequest(item), active)
		if err != nil {
			return forwardFailedMsg{Ref: item.ref(), Err: err.Error()}
		}
		if result.Err != nil {
			return forwardFailedMsg{Ref: item.ref(), Err: result.Err.Error()}
		}
		return forwardStartedMsg{Ref: item.ref(), SessionID: result.SessionID}
	}
}

func startForwardsCmd(ctx context.Context, svc appruntime.Service, selected []SelectedItem, active []domain.ForwardSession) tea.Cmd {
	requests := make([]domain.ForwardRequest, 0, len(selected))
	for _, item := range selected {
		requests = append(requests, selectedToRequest(item))
	}
	return func() tea.Msg {
		results, err := svc.StartMany(ctx, requests, active)
		msg := forwardBatchMsg{}
		if err != nil {
			for _, request := range requests {
				msg.failed = append(msg.failed, forwardFailedMsg{Ref: requestRef(request), Err: err.Error()})
			}
			return msg
		}
		for _, result := range results {
			if result.Err != nil {
				msg.failed = append(msg.failed, forwardFailedMsg{Ref: requestRef(result.Request), Err: result.Err.Error()})
				continue
			}
			msg.started = append(msg.started, forwardStartedMsg{Ref: requestRef(result.Request), SessionID: result.SessionID})
		}
		return msg
	}
}

func stopForwardCmd(ctx context.Context, runner ports.ForwardRunner, sessionID string) tea.Cmd {
	return func() tea.Msg {
		if err := runner.Stop(ctx, sessionID); err != nil {
			return forwardStopFailedMsg{SessionID: sessionID, Err: err.Error()}
		}
		return forwardStoppedMsg{SessionID: sessionID}
	}
}

func requestRef(request domain.ForwardRequest) forwardRef {
	return forwardRef{Context: request.Context, Namespace: request.Namespace, TargetID: request.TargetID}
}

func selectedToRequest(item SelectedItem) domain.ForwardRequest {
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
