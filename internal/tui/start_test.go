package tui

import (
	"context"
	"testing"

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

func TestStartKeyMovesSelectedItemsToRunningWithStartingStatus(t *testing.T) {
	m := NewModel(Dependencies{Runtime: &recordingRunner{}})
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
	m := NewModel(Dependencies{Runtime: runner})
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
