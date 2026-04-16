# Agent notes (port-forward-tui)

Compact context for coding agents. Prefer `README.md` for product behavior; this file is for **repo mechanics and pitfalls**.

## Module and layout

- **Go module**: `port-forward-tui` (`go.mod`). **Main binary**: `go run ./cmd/portfwd-tui` / `go build -o portfwd-tui ./cmd/portfwd-tui`.
- **Composition / wiring**: `cmd/portfwd-tui/main.go` bootstraps adapters (`internal/adapters/...`), app services (`internal/app/...`), and the Bubble Tea model (`internal/tui`).
- **Hex-ish split**: `internal/domain` (types), `internal/app/*` (catalog + runtime orchestration), `internal/ports` (interfaces), `internal/adapters` (kubectl, JSON config, `os/exec`), `internal/tui` (UI). `internal/tui/components` holds leaf views (often no `_test.go` there—tests tend to live next to the model in `internal/tui`).

## Commands (verified)

- **All tests**: `go test ./...`
- **Single package**: `go test ./internal/tui` (adjust path)
- **Single test**: `go test ./internal/tui -run '^TestName$' -count=1`
- **Static checks**: `go vet ./...` (no repo-local golangci-lint / Makefile / CI workflow in-tree as of last check—do not assume extra gates exist)

## Tests: README vs code

- `README.md` suggests `go test ./... -short` for “unit only” and full `./...` for integration with real `kubectl`. **The tree does not use `testing.Short()` anywhere**, so `-short` currently does **not** change which tests run.
- `test/integration/` exists, but its tests are **`t.Skip` placeholders** until wiring is finished—they are not opt-in kubectl integration tests yet.

## Runtime / env (easy to miss)

- **kubectl** must be on `PATH` for discovery and forwards (see `internal/adapters/kubectl`).
- **Config directory**: override with `PORTFWD_TUI_CONFIG_DIR`; otherwise uses OS user config dir + `portfwd-tui` (see `cmd/portfwd-tui/main.go`).

## Build artifacts

- Release-style builds often use `CGO_ENABLED=0` and `-trimpath -ldflags="-s -w"` (see `README.md`). `dist/` and `./portfwd-tui` are gitignored.
