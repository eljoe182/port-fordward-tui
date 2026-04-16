package tui

import "testing"

func TestRuntimeEventMarksForwardAsFailed(t *testing.T) {
	m := NewModel(Dependencies{})
	m.running = []RunningItem{{TargetID: "service:cco:admin", Status: StatusStarting}}

	next, _ := m.Update(RuntimeEvent{TargetID: "service:cco:admin", Status: StatusFailed, Err: "port in use"})
	updated := next.(Model)

	if updated.running[0].Status != StatusFailed || updated.running[0].Err != "port in use" {
		t.Fatalf("expected failed runtime state, got %#v", updated.running[0])
	}
}
