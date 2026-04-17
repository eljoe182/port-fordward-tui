# portfwd-tui

TUI en Go + Bubble Tea para gestionar `kubectl port-forward` sobre múltiples targets
(services y pods) con descubrimiento híbrido, persistencia local y runtime no destructivo.

## Requisitos

- Go 1.24+ (see `go.mod`)
- `kubectl` disponible en el `PATH`
- Acceso a un cluster Kubernetes con contexts configurados

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

## Atajos de teclado

| Tecla        | Acción                                                       |
| ------------ | ------------------------------------------------------------ |
| `↑` / `k`    | Cursor arriba en catálogo                                    |
| `↓` / `j`    | Cursor abajo en catálogo                                     |
| `Enter`      | Agregar target bajo el cursor a `Selected`                   |
| `f`          | Alternar favorito sobre el target bajo el cursor             |
| `c`          | Cambiar al siguiente contexto y recargar catálogo            |
| `n`          | Cambiar al siguiente namespace y recargar catálogo           |
| `r`          | Refrescar catálogo usando contexto/namespace actual          |
| `s`          | Iniciar port-forwards para todos los items en `Selected`     |
| `x`          | Detener el forward bajo el cursor en tab `Running`           |
| `Tab`        | Alternar entre tabs `Selected` / `Running`                   |
| `Esc`        | Limpiar el error actual del header                           |
| `q` / `Ctrl+C` | Salir con cleanup ordenado                                 |

## Persistencia

La configuración vive en JSON:

- Linux: `~/.config/portfwd-tui/config.json`
- macOS: `~/Library/Application Support/portfwd-tui/config.json`
- Windows: `%AppData%\portfwd-tui\config.json`
- Override: `PORTFWD_TUI_CONFIG_DIR=/ruta/custom`

La config guarda, por target, alias, puerto local preferido, favoritos, metadata mínima del target y recencia de uso.

## Arquitectura

```
cmd/portfwd-tui/          composition root
internal/domain/          Target, ForwardSession, AppConfig
internal/app/catalog/     merge + ranking Smart
internal/app/runtime/     validación + orquestación de forwards
internal/ports/           interfaces (Kubernetes, ConfigStore, ForwardRunner)
internal/adapters/        kubectl, configfile, exec (os/exec)
internal/tui/             Bubble Tea Model + Update + View
test/integration/         integración con kubectl real (opt-in)
```

## Tests

```bash
go test ./... -short     # unit tests
go test ./...            # incluye integration (kubectl real requerido)
```
