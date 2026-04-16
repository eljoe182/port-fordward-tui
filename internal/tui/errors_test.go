package tui

import "testing"

func TestActionableErrorMapsPortConflicts(t *testing.T) {
	got := actionableError("listen tcp 127.0.0.1:3001: bind: address already in use")
	if got != "local port unavailable — edit the local port and retry" {
		t.Fatalf("unexpected actionable error: %q", got)
	}
}

func TestActionableErrorMapsAuthorizationFailures(t *testing.T) {
	got := actionableError("Error from server (Forbidden): pods is forbidden")
	if got != "access denied — verify kubectl credentials and cluster permissions" {
		t.Fatalf("unexpected actionable error: %q", got)
	}
}
