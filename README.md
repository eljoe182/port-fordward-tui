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

| Tecla        | Acción                                   |
| ------------ | ---------------------------------------- |
| `Tab`        | Alternar entre tabs `Selected` / `Running` |
| `Enter`      | Agregar target del catálogo a `Selected` |
| `Ctrl+C`     | Salir con cleanup ordenado               |

> Atajos adicionales (navegación, edición de puerto, start/stop) se van cableando
> en tasks posteriores al MVP.

## Persistencia

La configuración vive en JSON:

- Linux: `~/.config/portfwd-tui/config.json`
- macOS: `~/Library/Application Support/portfwd-tui/config.json`
- Windows: `%AppData%\portfwd-tui\config.json`
- Override: `PORTFWD_TUI_CONFIG_DIR=/ruta/custom`

La config guarda, por target, alias, puerto local preferido y favoritos.

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
