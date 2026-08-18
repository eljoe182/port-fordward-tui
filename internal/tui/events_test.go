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
	m.running = []RunningItem{
		{TargetID: "service:admin", Context: "dev", SessionID: "sid-dev", Status: StatusRunning},
		{TargetID: "service:admin", Context: "prod", SessionID: "sid-prod", Status: StatusRunning},
	}

	next, _ := m.Update(forwardEventMsg{event: domain.ForwardEvent{
		SessionID: "sid-prod",
		TargetID:  "service:admin",
		Status:    domain.ForwardStatusFailed,
		Err:       "exit status 1",
	}})
	updated := next.(Model)

	if updated.running[0].Status != StatusRunning {
		t.Fatalf("expected dev session unchanged, got %+v", updated.running[0])
	}
	if updated.running[1].Status != StatusFailed || updated.running[1].Err != "exit status 1" {
		t.Fatalf("expected prod session failed, got %+v", updated.running[1])
	}
}

func TestForwardEventMsgRemovesRunningItemOnCleanStop(t *testing.T) {
	m := NewModel(Dependencies{})
	m.running = []RunningItem{
		{TargetID: "service:admin", Context: "dev", SessionID: "sid-dev", Status: StatusRunning},
		{TargetID: "service:admin", Context: "prod", SessionID: "sid-prod", Status: StatusRunning},
	}

	next, _ := m.Update(forwardEventMsg{event: domain.ForwardEvent{
		SessionID: "sid-dev",
		TargetID:  "service:admin",
		Status:    domain.ForwardStatusStopped,
	}})
	updated := next.(Model)

	if len(updated.running) != 1 || updated.running[0].SessionID != "sid-prod" {
		t.Fatalf("expected only dev session removed, got %#v", updated.running)
	}
}

func TestForwardEventWithUnknownSessionDoesNotMutateHomonymousRows(t *testing.T) {
	m := NewModel(Dependencies{})
	m.running = []RunningItem{
		{TargetID: "service:admin", Context: "dev", SessionID: "sid-dev", Status: StatusRunning},
		{TargetID: "service:admin", Context: "prod", SessionID: "sid-prod", Status: StatusRunning},
	}

	next, _ := m.Update(forwardEventMsg{event: domain.ForwardEvent{
		SessionID: "sid-unknown", TargetID: "service:admin", Status: domain.ForwardStatusFailed, Err: "boom",
	}})
	updated := next.(Model)

	for _, item := range updated.running {
		if item.Status != StatusRunning || item.Err != "" {
			t.Fatalf("expected unknown event to be ignored, got %#v", updated.running)
		}
	}
}

func TestForwardEventBeforeStartResultIsAppliedAfterSessionBinding(t *testing.T) {
	m := NewModel(Dependencies{})
	ref := forwardRef{Context: "prod", Namespace: "cco", TargetID: "service:cco:admin"}
	m.running = []RunningItem{{
		Context: ref.Context, Namespace: ref.Namespace, TargetID: ref.TargetID, Status: StatusStarting,
	}}

	next, _ := m.Update(forwardEventMsg{event: domain.ForwardEvent{
		SessionID: "sid-prod", TargetID: ref.TargetID, Status: domain.ForwardStatusFailed, Err: "exit status 1",
	}})
	m = next.(Model)
	next, _ = m.Update(forwardStartedMsg{Ref: ref, SessionID: "sid-prod"})
	updated := next.(Model)

	if updated.running[0].SessionID != "sid-prod" || updated.running[0].Status != StatusFailed {
		t.Fatalf("expected pending event applied after session binding, got %+v", updated.running[0])
	}
}

func TestListenForwardEventsCmdForwardsChannelEvents(t *testing.T) {
	runner := newChannelRunner()
	runner.events <- domain.ForwardEvent{SessionID: "sid-1", TargetID: "service:admin", Status: domain.ForwardStatusFailed, Err: "boom"}

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
		if msg.event.SessionID != "sid-1" || msg.event.TargetID != "service:admin" || msg.event.Status != domain.ForwardStatusFailed {
			t.Fatalf("event not piped correctly: %#v", msg.event)
		}
	case <-time.After(time.Second):
		t.Fatalf("cmd did not return within 1s")
	}
}
