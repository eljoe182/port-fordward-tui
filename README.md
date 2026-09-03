# portfwd-tui

Go + Bubble Tea TUI for managing `kubectl port-forward` across multiple targets
(services and pods) with hybrid discovery, local persistence, and a non-destructive runtime.

## Requirements

- Go 1.24+ (see `go.mod`)
- `kubectl` available on `PATH`
- Access to a Kubernetes cluster with configured contexts
- Optional: Azure CLI (`az`) on `PATH` and an authenticated session, for `Ctrl+R` AKS credential sync

## Install (Homebrew)

The GitHub release workflow publishes prebuilt archives and bumps `Formula/portfwd-tui.rb` on the default branch so this repository doubles as a Homebrew tap.

```bash
brew tap eljoe182/port-fordward-tui https://github.com/eljoe182/port-fordward-tui.git
brew update
brew install portfwd-tui
```

Use your own `OWNER/REPO` if you install from a fork. You need a published version tag (for example `v1.0.0`) so the formula’s URLs and checksums exist on the Releases page.

If the default branch is protected against direct pushes, add a repository secret named `HOMEBREW_FORMULA_PUSH_TOKEN` (PAT with `contents: write` on this repository) so the workflow can push the formula update.

## Build and run

Run from source (requires Go on the machine):

```bash
go run ./cmd/portfwd-tui
```

Build a binary in the current directory:

```bash
go build -o portfwd-tui ./cmd/portfwd-tui
chmod +x portfwd-tui   # Linux / macOS if needed
./portfwd-tui
```

Smaller release-style binary (static linking, strip debug info):

```bash
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o portfwd-tui ./cmd/portfwd-tui
```

Cross-compile examples (run from the repository root; adjust `GOOS` / `GOARCH` as needed):

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/portfwd-tui-linux-amd64 ./cmd/portfwd-tui
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/portfwd-tui-darwin-arm64 ./cmd/portfwd-tui
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/portfwd-tui-darwin-amd64 ./cmd/portfwd-tui
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/portfwd-tui-windows-amd64.exe ./cmd/portfwd-tui
```

End users only need the built executable plus `kubectl` and a valid cluster context; they do not need Go installed.

## Keyboard shortcuts

| Key            | Action                                                    |
| -------------- | --------------------------------------------------------- |
| `↑` / `k`      | Move the catalog cursor up                                |
| `↓` / `j`      | Move the catalog cursor down                              |
| `Enter`        | Add the target under the cursor to `Selected`             |
| `f`            | Toggle favorite for the target under the cursor           |
| `c`            | Open the context selector and reload the catalog          |
| `n`            | Open the namespace selector and reload the catalog        |
| `r`            | Refresh the catalog for the current context and namespace |
| `Ctrl+R`       | Sync Azure AKS credentials into kubeconfig, then refresh |
| `s`            | Start port-forwards for every item in `Selected`          |
| `x`            | Remove/stop in `Selected`; stop in `Running`               |
| `R`            | Retry the highlighted failed forward                      |
| `e`            | Edit the highlighted local port in `Selected`             |
| `J` / `K`      | Move the cursor within the active tab                     |
| `/`            | Search the catalog                                        |
| `t`            | Open the filter selector                                  |
| `o`            | Open the sort selector                                    |
| `Tab`          | Switch between the `Selected` and `Running` tabs          |
| `Esc`          | Clear the current header error                            |
| `q` / `Ctrl+C` | Exit with orderly cleanup                                 |

## Multi-context selection

`Selected` keeps its targets when you change context or namespace. Each row
stores and displays its source context, so you can select a target in `dev`,
switch to `prod`, and press `s` once to start both forwards. Active processes
are independent: stopping or retrying one does not affect a homonymous target
from another context.

Pressing `x` in `Selected` removes the highlighted target. If its matching
forward is starting or running, the TUI stops it first and removes both rows
after confirmation; a startup failure completes the removal automatically.
Pressing `x` in `Running` stops only that forward and keeps its selection
available for another start.

When you add a target, the TUI keeps its preferred port when that port is
available. If the port is already reserved by another selection, an active
forward, or another local process, the TUI chooses the next available port.
This fallback is temporary and does not replace the saved preference; explicit
changes made with `e` are persisted.

The local availability check is preventive. `kubectl` remains the final
authority because another process can claim the port between the check and
startup.

## Persistence

Configuration is stored as JSON:

- Linux: `~/.config/portfwd-tui/config.json`
- macOS: `~/Library/Application Support/portfwd-tui/config.json`
- Windows: `%AppData%\portfwd-tui\config.json`
- Override: `PORTFWD_TUI_CONFIG_DIR=/custom/path`

For each target, the configuration stores its alias, preferred local port,
favorite status, minimal metadata, and recent-use information.

## Architecture

```
cmd/portfwd-tui/          composition root
internal/domain/          Target, ForwardSession, AppConfig
internal/app/catalog/     merge + ranking Smart
internal/app/runtime/     validation + forward orchestration
internal/ports/           interfaces (Kubernetes, ConfigStore, ForwardRunner)
internal/adapters/        kubectl, configfile, exec (os/exec)
internal/tui/             Bubble Tea Model + Update + View
test/integration/         real kubectl integration (opt-in)
```

## Tests

```bash
go test ./...
go vet ./...
```

The tests under `test/integration/` are placeholders skipped with `t.Skip`. In
the current state, `-short` does not change which tests run.
