# Accumulated Multi-Context Selection Design

## Goal

Allow users to select targets while browsing different Kubernetes contexts and
start all selected port-forwards in one operation without losing each target's
original context or namespace.

## Scope

- The catalog continues to display one context and namespace at a time.
- Changing the catalog context preserves the current selection.
- Selected and running rows display their context and namespace.
- Homonymous targets in different contexts are independent.
- Preferred ports remain shared by the existing persisted target key.
- Automatically chosen fallback ports are session-only.

Combined multi-context discovery, presets, session restoration, and a config
schema migration are out of scope.

## Identity Model

The persisted `TargetID` remains unchanged. Before a process starts, the TUI
identifies a selection with the structured tuple `(context, namespace,
targetID)`. After a process starts, lifecycle events are correlated by the
runtime-generated `SessionID`.

This separation avoids changing `config.json` while allowing the same target
path to run in multiple contexts.

## Selection And Port Allocation

Each catalog item carries the resolved context. Selecting it copies context,
namespace, type, and target ID into `SelectedItem`.

The effective local port is chosen by scanning upward from the preferred port.
Ports already used by selected items, starting forwards, running forwards, or
another local process are skipped. A listener-based availability check is
advisory because another process can claim the port before `kubectl` starts.

Automatic fallback ports are not persisted. Explicit edits continue to update
the persisted preferred port.

## Start And Lifecycle Flow

Starting a batch creates one `ForwardRequest` per selected item using that
item's captured provenance. `StartMany` calls `cmd.Start()` for each request;
the calls return immediately and the resulting processes run concurrently.

Initial start results use the structured pre-start identity. Runtime exit and
stop events use `SessionID`. Events that arrive before the TUI binds a returned
session ID are held temporarily and applied immediately after binding.

Validation rejects every duplicate local port, regardless of target identity.
A per-request `kubectl` failure does not stop successful forwards in the same
batch.

## Persistence

The current `AppConfig` schema and target keys remain unchanged. Context,
selection, running sessions, and automatic fallback ports are transient.
Recency updates preserve an existing preferred port and do not overwrite the
currently browsed context with one selection's origin.

## Error Handling

- Port-range exhaustion leaves the target unselected and shows an actionable
  error.
- Batch validation failures transition optimistic rows out of `starting`.
- Stop waits for confirmation before removing a running row.
- Unknown or stale session events do not mutate another homonymous forward.
- `kubectl` remains authoritative for bind-time failures.

## Verification

Unit tests cover context-aware selection, provenance-preserving requests,
fallback port allocation, listener probing, strict duplicate-port validation,
session-isolated lifecycle events, persistence behavior, and row-level context
rendering. Final verification runs `go test ./...` and `go vet ./...`.
