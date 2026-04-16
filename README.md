# portfwd-tui

TUI en Go + Bubble Tea para gestionar `kubectl port-forward` sobre múltiples targets
(services y pods) con descubrimiento híbrido, persistencia local y runtime no destructivo.

## Requisitos

- Go 1.22+
- `kubectl` disponible en el `PATH`
- Acceso a un cluster Kubernetes con contexts configurados

## Uso

```bash
go run ./cmd/portfwd-tui
```

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
