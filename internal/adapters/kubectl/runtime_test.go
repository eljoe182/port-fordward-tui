package kubectl

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"cco-port-forward-tui/internal/domain"
)

func TestRuntimeEmitsFailedEventWhenProcessExitsUnexpectedly(t *testing.T) {
	runtime := NewRuntimeWithBuilder(func(ctx context.Context, _ domain.ForwardRequest) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "exit 3")
	})

	_, err := runtime.Start(context.Background(), domain.ForwardRequest{TargetID: "service:admin"})
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	event := waitEvent(t, runtime.Events())
	if event.Status != domain.ForwardStatusFailed {
		t.Fatalf("expected failed status, got %s", event.Status)
	}
	if event.TargetID != "service:admin" {
		t.Fatalf("expected target id preserved, got %q", event.TargetID)
	}
	if event.Err == "" {
		t.Fatalf("expected error message populated on failure")
	}
}

func TestRuntimeEmitsStoppedEventWhenStopCalledDeliberately(t *testing.T) {
	runtime := NewRuntimeWithBuilder(func(ctx context.Context, _ domain.ForwardRequest) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "sleep 5")
	})

	sessionID, err := runtime.Start(context.Background(), domain.ForwardRequest{TargetID: "service:admin"})
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	if err := runtime.Stop(context.Background(), sessionID); err != nil {
		t.Fatalf("stop failed: %v", err)
	}

	event := waitEvent(t, runtime.Events())
	if event.Status != domain.ForwardStatusStopped {
		t.Fatalf("expected stopped status, got %s (err=%q)", event.Status, event.Err)
	}
	if event.SessionID != sessionID {
		t.Fatalf("expected session id %q, got %q", sessionID, event.SessionID)
	}
}

func waitEvent(t *testing.T, ch <-chan domain.ForwardEvent) domain.ForwardEvent {
	t.Helper()
	select {
	case ev := <-ch:
		return ev
	case <-time.After(2 * time.Second):
		t.Fatalf("no event received within timeout")
		return domain.ForwardEvent{}
	}
}
