package tui

import (
	"context"
	"testing"
	"time"

	"port-forward-tui/internal/domain"
)

type channelRunner struct {
	events chan domain.ForwardEvent
}

func newChannelRunner() *channelRunner {
	return &channelRunner{events: make(chan domain.ForwardEvent, 4)}
}

func (c *channelRunner) Start(_ context.Context, req domain.ForwardRequest) (string, error) {
	return req.TargetID, nil
}
func (c *channelRunner) Stop(_ context.Context, _ string) error { return nil }
func (c *channelRunner) Events() <-chan domain.ForwardEvent     { return c.events }

func TestForwardEventMsgMarksRunningItemAsFailedAndStoresError(t *testing.T) {
	m := NewModel(Dependencies{})
	m.running = []RunningItem{{TargetID: "service:admin", Status: StatusRunning}}

	next, _ := m.Update(forwardEventMsg{event: domain.ForwardEvent{
		TargetID: "service:admin",
		Status:   domain.ForwardStatusFailed,
		Err:      "exit status 1",
	}})
	updated := next.(Model)

	if updated.running[0].Status != StatusFailed {
		t.Fatalf("expected failed status, got %s", updated.running[0].Status)
	}
	if updated.running[0].Err != "exit status 1" {
		t.Fatalf("expected error message stored, got %q", updated.running[0].Err)
	}
}

func TestForwardEventMsgRemovesRunningItemOnCleanStop(t *testing.T) {
	m := NewModel(Dependencies{})
	m.running = []RunningItem{
		{TargetID: "service:admin", Status: StatusRunning},
		{TargetID: "pod:redis", Status: StatusRunning},
	}

	next, _ := m.Update(forwardEventMsg{event: domain.ForwardEvent{
		TargetID: "service:admin",
		Status:   domain.ForwardStatusStopped,
	}})
	updated := next.(Model)

	if len(updated.running) != 1 || updated.running[0].TargetID != "pod:redis" {
		t.Fatalf("expected service:admin removed, got %#v", updated.running)
	}
}

func TestListenForwardEventsCmdForwardsChannelEvents(t *testing.T) {
	runner := newChannelRunner()
	runner.events <- domain.ForwardEvent{TargetID: "service:admin", Status: domain.ForwardStatusFailed, Err: "boom"}

	cmd := listenForwardEventsCmd(runner)
	if cmd == nil {
		t.Fatalf("expected cmd, got nil")
	}

	msgCh := make(chan any, 1)
	go func() { msgCh <- cmd() }()

	select {
	case got := <-msgCh:
		msg, ok := got.(forwardEventMsg)
		if !ok {
			t.Fatalf("expected forwardEventMsg, got %T", got)
		}
		if msg.event.TargetID != "service:admin" || msg.event.Status != domain.ForwardStatusFailed {
			t.Fatalf("event not piped correctly: %#v", msg.event)
		}
	case <-time.After(time.Second):
		t.Fatalf("cmd did not return within 1s")
	}
}
