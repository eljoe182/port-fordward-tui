package tui

import (
	"context"
	"testing"

	appruntime "port-forward-tui/internal/app/runtime"

	tea "github.com/charmbracelet/bubbletea"

	"port-forward-tui/internal/domain"
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
		{Context: "staging", Namespace: "payments", Type: "service", TargetID: "service:payments:admin", Label: "admin", LocalPort: 3001, RemotePort: 3000},
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	updated := next.(Model)

	if len(updated.running) != 1 {
		t.Fatalf("expected 1 running item, got %d", len(updated.running))
	}
	if updated.running[0].Status != StatusStarting {
		t.Fatalf("expected starting status, got %s", updated.running[0].Status)
	}
	if updated.running[0].Context != "staging" || updated.running[0].Namespace != "payments" || updated.running[0].Type != "service" {
		t.Fatalf("expected retry metadata persisted, got %+v", updated.running[0])
	}
	if updated.activeTab != TabRunning {
		t.Fatalf("expected switch to running tab, got %s", updated.activeTab)
	}
}

func TestStartForwardsUsesEachSelectedItemsProvenance(t *testing.T) {
	runner := &recordingRunner{}
	service := appruntime.NewService(runner)
	selected := []SelectedItem{
		{Context: "dev", Namespace: "cco", Type: "service", TargetID: "service:cco:admin", Label: "admin", LocalPort: 3001, RemotePort: 3000},
		{Context: "prod", Namespace: "cco", Type: "service", TargetID: "service:cco:admin", Label: "admin", LocalPort: 3002, RemotePort: 3000},
	}

	msg := startForwardsCmd(context.Background(), service, selected, nil)()
	if _, ok := msg.(forwardBatchMsg); !ok {
		t.Fatalf("expected forwardBatchMsg, got %T", msg)
	}
	if len(runner.starts) != 2 {
		t.Fatalf("expected two start requests, got %#v", runner.starts)
	}
	if runner.starts[0].Context != "dev" || runner.starts[1].Context != "prod" {
		t.Fatalf("expected per-selection contexts, got %#v", runner.starts)
	}
	if runner.starts[0].Namespace != "cco" || runner.starts[1].Namespace != "cco" {
		t.Fatalf("expected per-selection namespaces, got %#v", runner.starts)
	}
}

func TestHomonymousTargetInAnotherContextCanStart(t *testing.T) {
	runner := &recordingRunner{}
	m := NewModel(Dependencies{Runtime: runner, RuntimeApp: appruntime.NewService(runner)})
	m.running = []RunningItem{{
		Context: "dev", Namespace: "cco", Type: "service", TargetID: "service:cco:admin", LocalPort: 3001, Status: StatusRunning,
	}}
	m.selected = []SelectedItem{{
		Context: "prod", Namespace: "cco", Type: "service", TargetID: "service:cco:admin", LocalPort: 3002, RemotePort: 3000,
	}}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	updated := next.(Model)

	if len(updated.running) != 2 || updated.running[1].Context != "prod" {
		t.Fatalf("expected homonymous prod target to start, got %#v", updated.running)
	}
}

func TestForwardStartedMsgTransitionsRunningToRunningStatus(t *testing.T) {
	m := NewModel(Dependencies{})
	ref := forwardRef{Context: "dev", Namespace: "default", TargetID: "service:admin"}
	m.running = []RunningItem{{Context: ref.Context, Namespace: ref.Namespace, TargetID: ref.TargetID, Status: StatusStarting}}

	next, _ := m.Update(forwardStartedMsg{Ref: ref, SessionID: "service:admin"})
	updated := next.(Model)

	if updated.running[0].Status != StatusRunning {
		t.Fatalf("expected running status, got %s", updated.running[0].Status)
	}
	if updated.running[0].SessionID != "service:admin" {
		t.Fatalf("expected session id stored, got %q", updated.running[0].SessionID)
	}
}

func TestForwardBatchAssignsSessionsByContextAwareIdentity(t *testing.T) {
	m := NewModel(Dependencies{})
	dev := forwardRef{Context: "dev", Namespace: "cco", TargetID: "service:cco:admin"}
	prod := forwardRef{Context: "prod", Namespace: "cco", TargetID: "service:cco:admin"}
	m.running = []RunningItem{
		{Context: dev.Context, Namespace: dev.Namespace, TargetID: dev.TargetID, Status: StatusStarting},
		{Context: prod.Context, Namespace: prod.Namespace, TargetID: prod.TargetID, Status: StatusStarting},
	}

	next, _ := m.Update(forwardBatchMsg{started: []forwardStartedMsg{
		{Ref: prod, SessionID: "sid-prod"},
		{Ref: dev, SessionID: "sid-dev"},
	}})
	updated := next.(Model)

	if updated.running[0].SessionID != "sid-dev" || updated.running[1].SessionID != "sid-prod" {
		t.Fatalf("expected sessions assigned by context-aware identity, got %#v", updated.running)
	}
}

func TestBatchValidationFailureTransitionsEveryStartingRowToFailed(t *testing.T) {
	runner := &recordingRunner{}
	service := appruntime.NewService(runner)
	selected := []SelectedItem{
		{Context: "dev", Namespace: "cco", Type: "service", TargetID: "service:cco:admin", LocalPort: 3001, RemotePort: 3000},
		{Context: "prod", Namespace: "cco", Type: "service", TargetID: "service:cco:admin", LocalPort: 3001, RemotePort: 3000},
	}
	m := NewModel(Dependencies{})
	for _, item := range selected {
		m.running = append(m.running, RunningItem{
			Context: item.Context, Namespace: item.Namespace, Type: item.Type, TargetID: item.TargetID, LocalPort: item.LocalPort, Status: StatusStarting,
		})
	}

	msg := startForwardsCmd(context.Background(), service, selected, nil)()
	next, _ := m.Update(msg)
	updated := next.(Model)

	for _, item := range updated.running {
		if item.Status != StatusFailed || item.Err == "" {
			t.Fatalf("expected validation failure applied to all rows, got %#v", updated.running)
		}
	}
	if len(runner.starts) != 0 {
		t.Fatalf("expected validation before process start, got %#v", runner.starts)
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

	if len(updated.running) != 1 {
		t.Fatalf("expected running retained until stop succeeds, got %#v", updated.running)
	}
	if cmd == nil {
		t.Fatalf("expected stop command to be returned")
	}
	msg := cmd()
	if msg == nil {
		t.Fatalf("expected stop command to emit a message")
	}
	next, _ = updated.Update(msg)
	updated = next.(Model)
	if len(updated.running) != 0 {
		t.Fatalf("expected stopped session removed after confirmation, got %#v", updated.running)
	}
	if len(runner.stops) != 1 || runner.stops[0] != "sid-1" {
		t.Fatalf("expected Stop called with sid-1, got %#v", runner.stops)
	}
}

func TestStopHomonymousSessionLeavesSiblingRunning(t *testing.T) {
	runner := &recordingRunner{}
	m := NewModel(Dependencies{Runtime: runner, RuntimeApp: appruntime.NewService(runner)})
	m.activeTab = TabRunning
	m.running = []RunningItem{
		{TargetID: "service:cco:admin", Context: "dev", Namespace: "cco", SessionID: "sid-dev", Status: StatusRunning},
		{TargetID: "service:cco:admin", Context: "prod", Namespace: "cco", SessionID: "sid-prod", Status: StatusRunning},
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	updated := next.(Model)
	msg := cmd()
	next, _ = updated.Update(msg)
	updated = next.(Model)

	if len(updated.running) != 1 || updated.running[0].SessionID != "sid-prod" {
		t.Fatalf("expected prod sibling to remain, got %#v", updated.running)
	}
	if len(runner.stops) != 1 || runner.stops[0] != "sid-dev" {
		t.Fatalf("expected only dev session stopped, got %#v", runner.stops)
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

func TestRetryHomonymousForwardOnlyUpdatesSelectedContext(t *testing.T) {
	runner := &recordingRunner{}
	m := NewModel(Dependencies{Runtime: runner, RuntimeApp: appruntime.NewService(runner)})
	m.activeTab = TabRunning
	m.running = []RunningItem{
		{TargetID: "service:cco:admin", Context: "dev", Namespace: "cco", Type: "service", SessionID: "sid-dev-old", LocalPort: 3001, RemotePort: 3000, Status: StatusFailed},
		{TargetID: "service:cco:admin", Context: "prod", Namespace: "cco", Type: "service", SessionID: "sid-prod", LocalPort: 3002, RemotePort: 3000, Status: StatusRunning},
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("R")})
	updated := next.(Model)
	msg := cmd()
	next, _ = updated.Update(msg)
	updated = next.(Model)

	if updated.running[0].Status != StatusRunning {
		t.Fatalf("expected dev retry to start, got %+v", updated.running[0])
	}
	if updated.running[1].Status != StatusRunning || updated.running[1].SessionID != "sid-prod" {
		t.Fatalf("expected prod sibling unchanged, got %+v", updated.running[1])
	}
}
