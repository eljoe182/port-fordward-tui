package tui

import (
	"context"
	"errors"
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	appruntime "port-forward-tui/internal/app/runtime"
	"port-forward-tui/internal/domain"
)

const removalTargetID = "service:cco:admin"

type removalRunner struct {
	stops    []string
	stopErr  error
	stopErrs []error
}

func (r *removalRunner) Start(_ context.Context, req domain.ForwardRequest) (string, error) {
	return req.TargetID, nil
}

func (r *removalRunner) Stop(_ context.Context, sessionID string) error {
	r.stops = append(r.stops, sessionID)
	if len(r.stopErrs) >= len(r.stops) {
		return r.stopErrs[len(r.stops)-1]
	}
	return r.stopErr
}

func (r *removalRunner) Events() <-chan domain.ForwardEvent { return nil }

func removalRef(contextName string) forwardRef {
	return forwardRef{Context: contextName, Namespace: "cco", TargetID: removalTargetID}
}

func removalSelected(ref forwardRef) SelectedItem {
	return SelectedItem{
		Context: ref.Context, Namespace: ref.Namespace, Type: "service", TargetID: ref.TargetID,
		LocalPort: 3001, RemotePort: 3000,
	}
}

func removalRunning(ref forwardRef, sessionID string, status ForwardStatus) RunningItem {
	return RunningItem{
		Context: ref.Context, Namespace: ref.Namespace, Type: "service", TargetID: ref.TargetID,
		SessionID: sessionID, LocalPort: 3001, RemotePort: 3000, Status: status,
	}
}

func pressRemovalKey(t *testing.T, m Model) (Model, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	return next.(Model), cmd
}

func applyRemovalCmd(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected command")
	}
	next, _ := m.Update(cmd())
	return next.(Model)
}

func selectedIDs(items []SelectedItem) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.TargetID)
	}
	if len(ids) == 0 {
		return nil
	}
	return ids
}

func TestRemoveSelectedUpdatesItemsAndCursor(t *testing.T) {
	tests := []struct {
		name       string
		items      []SelectedItem
		cursor     int
		wantIDs    []string
		wantCursor int
	}{
		{name: "first", items: []SelectedItem{{TargetID: "a"}, {TargetID: "b"}, {TargetID: "c"}}, wantIDs: []string{"b", "c"}},
		{name: "middle", items: []SelectedItem{{TargetID: "a"}, {TargetID: "b"}, {TargetID: "c"}}, cursor: 1, wantIDs: []string{"a", "c"}, wantCursor: 1},
		{name: "last", items: []SelectedItem{{TargetID: "a"}, {TargetID: "b"}, {TargetID: "c"}}, cursor: 2, wantIDs: []string{"a", "b"}, wantCursor: 1},
		{name: "only", items: []SelectedItem{{TargetID: "a"}}, wantIDs: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(Dependencies{})
			m.activeTab = TabSelected
			m.selected = tt.items
			m.selectedCursor = tt.cursor

			updated, cmd := pressRemovalKey(t, m)

			if cmd != nil {
				t.Fatal("expected removal without a running forward to be immediate")
			}
			if got := selectedIDs(updated.selected); !reflect.DeepEqual(got, tt.wantIDs) {
				t.Fatalf("expected selected IDs %v, got %v", tt.wantIDs, got)
			}
			if updated.selectedCursor != tt.wantCursor {
				t.Fatalf("expected cursor %d, got %d", tt.wantCursor, updated.selectedCursor)
			}
		})
	}
}

func TestRemoveSelectedStopsOnlyMatchingContext(t *testing.T) {
	runner := &removalRunner{}
	dev, prod := removalRef("dev"), removalRef("prod")
	m := NewModel(Dependencies{Runtime: runner})
	m.activeTab = TabSelected
	m.selected = []SelectedItem{removalSelected(dev), removalSelected(prod)}
	m.running = []RunningItem{
		removalRunning(dev, "sid-dev", StatusRunning),
		removalRunning(prod, "sid-prod", StatusRunning),
	}

	updated, cmd := pressRemovalKey(t, m)
	if len(updated.selected) != 2 || len(updated.running) != 2 {
		t.Fatal("expected rows retained until stop succeeds")
	}
	updated = applyRemovalCmd(t, updated, cmd)

	if len(updated.selected) != 1 || updated.selected[0].Context != "prod" {
		t.Fatalf("expected only prod selection to remain, got %#v", updated.selected)
	}
	if len(updated.running) != 1 || updated.running[0].SessionID != "sid-prod" {
		t.Fatalf("expected only prod forward to remain, got %#v", updated.running)
	}
	if !reflect.DeepEqual(runner.stops, []string{"sid-dev"}) {
		t.Fatalf("expected only sid-dev stopped, got %v", runner.stops)
	}
}

func TestRemoveSelectedWaitsForStartingForward(t *testing.T) {
	runner := &removalRunner{}
	ref := removalRef("dev")
	m := NewModel(Dependencies{Runtime: runner})
	m.activeTab = TabSelected
	m.selected = []SelectedItem{removalSelected(ref)}
	m.running = []RunningItem{removalRunning(ref, "", StatusStarting)}

	updated, cmd := pressRemovalKey(t, m)
	if cmd != nil {
		t.Fatal("expected removal to wait for a session ID")
	}
	next, stopCmd := updated.Update(forwardStartedMsg{Ref: ref, SessionID: "sid-dev"})
	updated = applyRemovalCmd(t, next.(Model), stopCmd)

	if len(updated.selected) != 0 || len(updated.running) != 0 {
		t.Fatalf("expected pending removal completed, selected=%#v running=%#v", updated.selected, updated.running)
	}
}

func TestPendingSelectedRemovalHandlesStartupFailures(t *testing.T) {
	setup := func() (Model, forwardRef) {
		ref := removalRef("dev")
		m := NewModel(Dependencies{})
		m.activeTab = TabSelected
		m.selected = []SelectedItem{removalSelected(ref)}
		m.running = []RunningItem{removalRunning(ref, "", StatusStarting)}
		m, _ = pressRemovalKey(t, m)
		return m, ref
	}

	t.Run("start result fails", func(t *testing.T) {
		m, ref := setup()
		next, _ := m.Update(forwardFailedMsg{Ref: ref, Err: "start failed"})
		updated := next.(Model)
		if len(updated.selected) != 0 || len(updated.running) != 0 {
			t.Fatalf("expected failed start to complete removal, selected=%#v running=%#v", updated.selected, updated.running)
		}
	})

	t.Run("process fails before session binding", func(t *testing.T) {
		m, ref := setup()
		next, _ := m.Update(forwardEventMsg{event: domain.ForwardEvent{
			SessionID: "sid-dev", TargetID: ref.TargetID, Status: domain.ForwardStatusFailed, Err: "exit status 1",
		}})
		m = next.(Model)
		next, cmd := m.Update(forwardStartedMsg{Ref: ref, SessionID: "sid-dev"})
		updated := next.(Model)
		if cmd != nil || len(updated.selected) != 0 || len(updated.running) != 0 {
			t.Fatalf("expected buffered failure to complete removal, selected=%#v running=%#v", updated.selected, updated.running)
		}
	})
}

func TestPendingSelectedRemovalHandlesStopEventBeforeCommandResult(t *testing.T) {
	runner := &removalRunner{}
	ref := removalRef("dev")
	m := NewModel(Dependencies{Runtime: runner})
	m.activeTab = TabSelected
	m.selected = []SelectedItem{removalSelected(ref)}
	m.running = []RunningItem{removalRunning(ref, "sid-dev", StatusRunning)}

	updated, cmd := pressRemovalKey(t, m)
	next, _ := updated.Update(forwardEventMsg{event: domain.ForwardEvent{
		SessionID: "sid-dev", TargetID: ref.TargetID, Status: domain.ForwardStatusStopped,
	}})
	updated = next.(Model)
	updated = applyRemovalCmd(t, updated, cmd)

	if len(updated.selected) != 0 || len(updated.running) != 0 {
		t.Fatalf("expected both stop paths to be idempotent, selected=%#v running=%#v", updated.selected, updated.running)
	}
}

func TestRemoveSelectedDropsFailedForwardWithoutStopping(t *testing.T) {
	runner := &removalRunner{}
	ref := removalRef("dev")
	m := NewModel(Dependencies{Runtime: runner})
	m.activeTab = TabSelected
	m.selected = []SelectedItem{removalSelected(ref)}
	m.running = []RunningItem{removalRunning(ref, "", StatusFailed)}

	updated, cmd := pressRemovalKey(t, m)

	if cmd != nil || len(updated.selected) != 0 || len(updated.running) != 0 || len(runner.stops) != 0 {
		t.Fatalf("expected immediate failed-row removal, selected=%#v running=%#v stops=%v", updated.selected, updated.running, runner.stops)
	}
}

func TestRemoveSelectedStopFailureIsVisibleAndRetryable(t *testing.T) {
	runner := &removalRunner{stopErr: errors.New("permission denied")}
	ref := removalRef("dev")
	m := NewModel(Dependencies{Runtime: runner})
	m.activeTab = TabSelected
	m.selected = []SelectedItem{removalSelected(ref)}
	m.running = []RunningItem{removalRunning(ref, "sid-dev", StatusRunning)}

	updated, cmd := pressRemovalKey(t, m)
	updated = applyRemovalCmd(t, updated, cmd)
	if len(updated.selected) != 1 || len(updated.running) != 1 || updated.errMsg == "" || updated.running[0].Err == "" {
		t.Fatalf("expected rows and visible error retained, selected=%#v running=%#v header=%q", updated.selected, updated.running, updated.errMsg)
	}
	_, retryCmd := pressRemovalKey(t, updated)
	if retryCmd == nil {
		t.Fatal("expected removal to be retryable")
	}
}

func TestPendingRemovalPreservesCursorTargets(t *testing.T) {
	runner := &removalRunner{}
	refs := []forwardRef{{TargetID: "a"}, {TargetID: "b"}, {TargetID: "c"}}
	m := NewModel(Dependencies{Runtime: runner})
	m.activeTab = TabSelected
	for i, ref := range refs {
		m.selected = append(m.selected, removalSelected(ref))
		m.running = append(m.running, removalRunning(ref, "sid-"+string(rune('a'+i)), StatusRunning))
	}

	updated, cmd := pressRemovalKey(t, m)
	updated.selectedCursor = 2
	updated.runningCursor = 2
	updated = applyRemovalCmd(t, updated, cmd)

	if updated.selectedCursor != 1 || updated.selected[1].TargetID != "c" {
		t.Fatalf("expected selected cursor to stay on c, cursor=%d selected=%#v", updated.selectedCursor, updated.selected)
	}
	if updated.runningCursor != 1 || updated.running[1].TargetID != "c" {
		t.Fatalf("expected running cursor to stay on c, cursor=%d running=%#v", updated.runningCursor, updated.running)
	}
}

func TestStopResultsRemainBoundToTheirIntent(t *testing.T) {
	runner := &removalRunner{stopErrs: []error{errors.New("first stop failed"), nil}}
	ref := removalRef("dev")
	m := NewModel(Dependencies{Runtime: runner})
	m.activeTab = TabRunning
	m.selected = []SelectedItem{removalSelected(ref)}
	m.running = []RunningItem{removalRunning(ref, "sid-dev", StatusRunning)}

	updated, runningStopCmd := pressRemovalKey(t, m)
	updated.activeTab = TabSelected
	updated, selectedStopCmd := pressRemovalKey(t, updated)
	updated = applyRemovalCmd(t, updated, runningStopCmd)
	updated.errMsg = "another operation failed"
	updated = applyRemovalCmd(t, updated, selectedStopCmd)

	if len(updated.selected) != 0 || len(updated.running) != 0 {
		t.Fatalf("expected selected removal to survive unrelated stop failure, selected=%#v running=%#v", updated.selected, updated.running)
	}
	if updated.errMsg != "another operation failed" {
		t.Fatalf("expected unrelated error preserved, got %q", updated.errMsg)
	}
}

func TestStopFromRunningKeepsSelectionAndCursorTarget(t *testing.T) {
	runner := &removalRunner{}
	ref := removalRef("dev")
	m := NewModel(Dependencies{Runtime: runner})
	m.activeTab = TabRunning
	m.selected = []SelectedItem{removalSelected(ref)}
	m.running = []RunningItem{
		removalRunning(ref, "sid-a", StatusRunning),
		{TargetID: "b", SessionID: "sid-b", Status: StatusRunning},
		{TargetID: "c", SessionID: "sid-c", Status: StatusRunning},
	}

	updated, cmd := pressRemovalKey(t, m)
	updated.runningCursor = 1
	updated = applyRemovalCmd(t, updated, cmd)

	if len(updated.selected) != 1 {
		t.Fatalf("expected selection retained, got %#v", updated.selected)
	}
	if updated.runningCursor != 0 || updated.running[0].TargetID != "b" {
		t.Fatalf("expected cursor to stay on b, cursor=%d running=%#v", updated.runningCursor, updated.running)
	}
}

func TestRemovalDuringRetryUsesNewSession(t *testing.T) {
	t.Run("Running does not stop previous session", func(t *testing.T) {
		runner := &removalRunner{}
		ref := removalRef("dev")
		m := NewModel(Dependencies{Runtime: runner, RuntimeApp: appruntime.NewService(runner)})
		m.activeTab = TabRunning
		m.running = []RunningItem{removalRunning(ref, "sid-old", StatusFailed)}

		next, retryCmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("R")})
		updated := next.(Model)
		if retryCmd == nil || updated.running[0].SessionID != "" {
			t.Fatalf("expected retry with cleared session, got %#v", updated.running)
		}
		_, stopCmd := pressRemovalKey(t, updated)
		if stopCmd != nil || len(runner.stops) != 0 {
			t.Fatalf("expected no stop while retry starts, stops=%v", runner.stops)
		}
	})

	t.Run("Selected stops new retry session", func(t *testing.T) {
		runner := &removalRunner{}
		ref := removalRef("dev")
		m := NewModel(Dependencies{Runtime: runner, RuntimeApp: appruntime.NewService(runner)})
		m.activeTab = TabRunning
		m.selected = []SelectedItem{removalSelected(ref)}
		m.running = []RunningItem{removalRunning(ref, "sid-old", StatusFailed)}

		next, retryCmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("R")})
		m = next.(Model)
		m.activeTab = TabSelected
		m, removeCmd := pressRemovalKey(t, m)
		if removeCmd != nil {
			t.Fatal("expected removal to wait for retry session")
		}
		next, stopCmd := m.Update(retryCmd())
		updated := applyRemovalCmd(t, next.(Model), stopCmd)

		if len(updated.selected) != 0 || len(updated.running) != 0 {
			t.Fatalf("expected retry removal completed, selected=%#v running=%#v", updated.selected, updated.running)
		}
		if !reflect.DeepEqual(runner.stops, []string{ref.TargetID}) {
			t.Fatalf("expected new session stopped, got %v", runner.stops)
		}
	})
}
