package kubectl

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"port-forward-tui/internal/domain"
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

func TestDefaultCommandBuilderUsesTargetNameFromNamespacedKey(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "args.txt")
	script := filepath.Join(dir, "kubectl")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf '%s\n' \"$@\" > \"$PORTFWD_CAPTURE\"\n"), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("PORTFWD_CAPTURE", output)

	cmd := defaultCommandBuilder(context.Background(), domain.ForwardRequest{
		TargetID:   "service:cco:admin",
		Context:    "dev",
		Namespace:  "cco",
		LocalPort:  3001,
		RemotePort: 3000,
		Type:       domain.TargetTypeService,
	})

	if err := cmd.Run(); err != nil {
		t.Fatalf("run command: %v", err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	args := string(data)
	if !strings.Contains(args, "service/admin") {
		t.Fatalf("expected resource name without namespace duplication, got %q", args)
	}
}
