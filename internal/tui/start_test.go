package tui

import (
	"context"
	"testing"

	appruntime "cco-port-forward-tui/internal/app/runtime"
	tea "github.com/charmbracelet/bubbletea"

	"cco-port-forward-tui/internal/domain"
)

type recordingRunner struct {
	starts []domain.ForwardRequest
	stops  []string
	err    error
}

func (r *recordingRunner) Start(_ context.Context, req domain.ForwardRequest) (string, error) {
	r.starts = append(r.starts, req)
	return req.TargetID, r.err
}
func (r *recordingRunner) Stop(_ context.Context, sessionID string) error {
	r.stops = append(r.stops, sessionID)
	return nil
}
func (r *recordingRunner) Events() <-chan domain.ForwardEvent { return nil }

func TestStartKeyMovesSelectedItemsToRunningWithStartingStatus(t *testing.T) {
	runner := &recordingRunner{}
	m := NewModel(Dependencies{Runtime: runner, RuntimeApp: appruntime.NewService(runner)})
	m.contextName = "dev"
	m.namespace = "cco"
	m.selected = []SelectedItem{
		{TargetID: "service:admin", Label: "admin", LocalPort: 3001, RemotePort: 3000},
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	updated := next.(Model)

	if len(updated.running) != 1 {
		t.Fatalf("expected 1 running item, got %d", len(updated.running))
	}
	if updated.running[0].Status != StatusStarting {
		t.Fatalf("expected starting status, got %s", updated.running[0].Status)
	}
	if updated.running[0].Context != "dev" || updated.running[0].Namespace != "cco" || updated.running[0].Type != "service" {
		t.Fatalf("expected retry metadata persisted, got %+v", updated.running[0])
	}
	if updated.activeTab != TabRunning {
		t.Fatalf("expected switch to running tab, got %s", updated.activeTab)
	}
}

func TestForwardStartedMsgTransitionsRunningToRunningStatus(t *testing.T) {
	m := NewModel(Dependencies{})
	m.running = []RunningItem{{TargetID: "service:admin", Status: StatusStarting}}

	next, _ := m.Update(forwardStartedMsg{TargetID: "service:admin", SessionID: "service:admin"})
	updated := next.(Model)

	if updated.running[0].Status != StatusRunning {
		t.Fatalf("expected running status, got %s", updated.running[0].Status)
	}
	if updated.running[0].SessionID != "service:admin" {
		t.Fatalf("expected session id stored, got %q", updated.running[0].SessionID)
	}
}

func TestStopKeyOnRunningTabRemovesForwardAndInvokesRunner(t *testing.T) {
	runner := &recordingRunner{}
	m := NewModel(Dependencies{Runtime: runner, RuntimeApp: appruntime.NewService(runner)})
	m.activeTab = TabRunning
	m.running = []RunningItem{
		{TargetID: "service:admin", SessionID: "sid-1", Status: StatusRunning},
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	updated := next.(Model)

	if len(updated.running) != 0 {
		t.Fatalf("expected running cleared, got %#v", updated.running)
	}
	if cmd == nil {
		t.Fatalf("expected stop command to be returned")
	}
	if msg := cmd(); msg == nil {
		t.Fatalf("expected stop command to emit a message")
	}
	if len(runner.stops) != 1 || runner.stops[0] != "sid-1" {
		t.Fatalf("expected Stop called with sid-1, got %#v", runner.stops)
	}
}

func TestRetryKeyRestartsFailedRunningItem(t *testing.T) {
	runner := &recordingRunner{}
	m := NewModel(Dependencies{Runtime: runner, RuntimeApp: appruntime.NewService(runner)})
	m.activeTab = TabRunning
	m.running = []RunningItem{{
		TargetID:   "service:cco:admin",
		Context:    "dev",
		Namespace:  "cco",
		Type:       "service",
		Label:      "admin",
		LocalPort:  3001,
		RemotePort: 3000,
		Status:     StatusFailed,
		Err:        "local port unavailable — edit the local port and retry",
	}}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("R")})
	updated := next.(Model)
	if updated.running[0].Status != StatusStarting || updated.running[0].Err != "" {
		t.Fatalf("expected retry to reset item state, got %+v", updated.running[0])
	}
	if cmd == nil {
		t.Fatalf("expected retry command")
	}
	msg := cmd()
	if _, ok := msg.(forwardStartedMsg); !ok {
		t.Fatalf("expected forwardStartedMsg, got %T", msg)
	}
	if len(runner.starts) != 1 || runner.starts[0].Namespace != "cco" || runner.starts[0].Context != "dev" {
		t.Fatalf("expected retry to use original request metadata, got %+v", runner.starts)
	}
}
