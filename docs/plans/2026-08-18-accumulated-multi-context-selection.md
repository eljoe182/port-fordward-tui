# Accumulated Multi-Context Selection Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Accumulate targets from different Kubernetes contexts and start them together with independent ports and lifecycle state.

**Architecture:** Keep the catalog single-context and attach structured context provenance to transient TUI items. Use a context-aware reference before startup, `SessionID` after startup, and a listener-backed port checker for best-effort local allocation without changing persisted target keys.

**Tech Stack:** Go 1.24, Bubble Tea, Lip Gloss, `kubectl`, Go standard library networking and testing.

---

### Task 1: Add context-aware selection identity

**Files:** `internal/tui/state.go`, `internal/tui/loader.go`, `internal/tui/model.go`, and their focused tests.

Capture context, namespace, and target type when selecting. Prevent duplicates only when the complete structured reference matches and preserve selections across catalog reloads.

### Task 2: Allocate available local ports

**Files:** `internal/ports/local_port.go`, `internal/adapters/localport/checker.go`, `internal/app/runtime/local_port.go`, `cmd/portfwd-tui/main.go`, and focused tests.

Scan from the preferred port while skipping selected, active, and host-occupied ports. Treat host probing as advisory and leave fallback ports transient.

### Task 3: Start requests with captured provenance

**Files:** `internal/tui/runtime.go`, `internal/tui/update.go`, `internal/app/runtime/service.go`, and focused tests.

Build every request and optimistic running row from its selected item. Snapshot pre-existing active sessions before inserting new rows and reject every duplicate local port.

### Task 4: Correlate process lifecycle by session

**Files:** `internal/tui/events.go`, `internal/tui/update.go`, `internal/tui/runtime.go`, `internal/tui/model.go`, and focused tests.

Use the structured reference for initial results and `SessionID` for process events and stops. Buffer events that arrive before session binding and avoid optimistic removal on stop.

### Task 5: Preserve persistence semantics and expose provenance

**Files:** `internal/tui/preferences.go`, `internal/tui/components/selected_tab.go`, `internal/tui/components/running_tab.go`, `internal/tui/view.go`, `README.md`, and focused tests.

Do not persist automatic fallback ports, keep manual edits persistent, avoid changing catalog context during recency writes, and render context/namespace on selected and running rows.

### Task 6: Verify the complete feature

Run focused package tests, `go test ./...`, and `go vet ./...`. With real Kubernetes access, select targets from two contexts, start both in one batch, and verify stopping one session leaves the other running.
