# Kubernetes Port-Forward TUI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Construir una TUI en Go + Bubble Tea que replique el MVP del script actual y añada descubrimiento híbrido, persistencia útil y runtime no destructivo.

**Architecture:** La implementación usará una arquitectura limpia y ligera: dominio simple para `Target`, `ForwardSession` y reglas de selección; capa de aplicación para catálogo, ranking y validación; adaptadores para `kubectl`, persistencia local y runtime de procesos; y una capa TUI con Bubble Tea desacoplada de la lógica operativa. El runtime de port-forward y el descubrimiento NO deben depender del renderizado para poder probarse con tests unitarios y de integración acotados.

**Tech Stack:** Go, Bubble Tea, Bubbles, Lip Gloss, os/exec, contexto/cancelación, JSON/YAML para persistencia local, testing estándar de Go.

---

## Architecture Blueprint

### Proposed package layout

```text
cmd/portfwd-tui/main.go
internal/domain/
  target.go
  forward_session.go
  config.go
internal/app/catalog/
  service.go
  ranking.go
  merge.go
  validation.go
internal/app/runtime/
  service.go
internal/ports/
  kubernetes.go
  runtime.go
  config_store.go
internal/adapters/kubectl/
  discovery.go
  runtime.go
internal/adapters/configfile/
  store.go
internal/tui/
  model.go
  update.go
  view.go
  keymap.go
  state.go
internal/tui/components/
  catalog.go
  selected_tab.go
  running_tab.go
  header.go
test/integration/
  kubectl_discovery_test.go
  runtime_service_test.go
```

### Core design decisions

1. **Un solo modelo de negocio para catálogo:** `Target` unifica `pod` y `service`.
2. **Ranking inteligente en aplicación, no en UI:** la TUI solo consume una lista ya resuelta.
3. **Runtime como servicio independiente:** la TUI emite comandos (`start`, `stop`, `retry`) y reacciona a eventos.
4. **Persistencia local simple:** usar JSON o YAML en `~/.config/portfwd-tui/config.json`.
5. **Sin acoplar tests a Bubble Tea:** probar reglas, merge, ranking y runtime fuera de la UI.

### Suggested startup flow

1. Cargar config local.
2. Leer current context desde `kubectl`.
3. Listar contexts y namespaces.
4. Descubrir targets del namespace actual.
5. Mezclar descubiertos + configurados.
6. Ordenar por `Smart`.
7. Renderizar layout con catálogo central y panel derecho con tabs.

---

## Implementation Tasks

### Task 1: Bootstrap del proyecto Go y composition root

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `cmd/portfwd-tui/main.go`
- Create: `internal/tui/model.go`
- Create: `internal/tui/update.go`
- Create: `internal/tui/view.go`
- Test: `internal/tui/model_test.go`

**Step 1: Write the failing test**

```go
package tui

import "testing"

func TestNewModelStartsWithSelectedTabAndEmptyState(t *testing.T) {
	model := NewModel(Dependencies{})

	if model.activeTab != TabSelected {
		t.Fatalf("expected default tab %q, got %q", TabSelected, model.activeTab)
	}

	if len(model.catalog) != 0 {
		t.Fatalf("expected empty catalog on startup")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestNewModelStartsWithSelectedTabAndEmptyState -v`
Expected: FAIL because `NewModel`, `Dependencies` o `TabSelected` no existen.

**Step 3: Write minimal implementation**

```go
package tui

type Tab string

const TabSelected Tab = "selected"

type Dependencies struct{}

type Model struct {
	activeTab Tab
	catalog   []string
}

func NewModel(_ Dependencies) Model {
	return Model{activeTab: TabSelected, catalog: []string{}}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui -run TestNewModelStartsWithSelectedTabAndEmptyState -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add go.mod go.sum cmd/portfwd-tui/main.go internal/tui/model.go internal/tui/update.go internal/tui/view.go internal/tui/model_test.go
git commit -m "chore: bootstrap go tui project"
```

---

### Task 2: Modelar dominio base para targets, sesiones y configuración

**Files:**
- Create: `internal/domain/target.go`
- Create: `internal/domain/forward_session.go`
- Create: `internal/domain/config.go`
- Test: `internal/domain/target_test.go`

**Step 1: Write the failing test**

```go
package domain

import "testing"

func TestTargetMergeConfiguredFieldsOverridesDiscoveredDefaults(t *testing.T) {
	discovered := Target{Name: "cco-admin-api", Type: TargetTypeService, RemotePort: 3000}
	configured := TargetConfig{Alias: "admin", PreferredLocalPort: 3001, Favorite: true}

	merged := discovered.WithConfig(configured)

	if merged.Alias != "admin" || merged.PreferredLocalPort != 3001 || !merged.Favorite {
		t.Fatalf("expected configured values to override discovered defaults: %+v", merged)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/domain -run TestTargetMergeConfiguredFieldsOverridesDiscoveredDefaults -v`
Expected: FAIL because `Target`, `TargetConfig` o `WithConfig` no existen.

**Step 3: Write minimal implementation**

```go
package domain

type TargetType string

const (
	TargetTypeService TargetType = "service"
	TargetTypePod     TargetType = "pod"
)

type Target struct {
	Name               string
	Type               TargetType
	Alias              string
	RemotePort         int
	PreferredLocalPort int
	Favorite           bool
}

type TargetConfig struct {
	Alias              string `json:"alias"`
	PreferredLocalPort int    `json:"preferredLocalPort"`
	Favorite           bool   `json:"favorite"`
}

func (t Target) WithConfig(cfg TargetConfig) Target {
	if cfg.Alias != "" {
		t.Alias = cfg.Alias
	}
	if cfg.PreferredLocalPort != 0 {
		t.PreferredLocalPort = cfg.PreferredLocalPort
	}
	t.Favorite = cfg.Favorite
	return t
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/domain -run TestTargetMergeConfiguredFieldsOverridesDiscoveredDefaults -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/domain/target.go internal/domain/forward_session.go internal/domain/config.go internal/domain/target_test.go
git commit -m "feat: add core domain models"
```

---

### Task 3: Definir puertos y adaptador de descubrimiento kubectl

**Files:**
- Create: `internal/ports/kubernetes.go`
- Create: `internal/adapters/kubectl/discovery.go`
- Test: `internal/adapters/kubectl/discovery_test.go`

**Step 1: Write the failing test**

```go
package kubectl

import (
	"context"
	"testing"
)

func TestListContextsParsesKubectlOutput(t *testing.T) {
	exec := fakeExec{"dev\nprod\n"}
	client := NewDiscoveryClient(exec)

	contexts, err := client.ListContexts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(contexts) != 2 || contexts[0] != "dev" || contexts[1] != "prod" {
		t.Fatalf("unexpected contexts: %#v", contexts)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/kubectl -run TestListContextsParsesKubectlOutput -v`
Expected: FAIL because `NewDiscoveryClient` y el contrato de ejecución aún no existen.

**Step 3: Write minimal implementation**

```go
package kubectl

import (
	"context"
	"strings"
)

type ExecRunner interface {
	Run(context.Context, string, ...string) (string, error)
}

type DiscoveryClient struct{ exec ExecRunner }

func NewDiscoveryClient(exec ExecRunner) DiscoveryClient {
	return DiscoveryClient{exec: exec}
}

func (c DiscoveryClient) ListContexts(ctx context.Context) ([]string, error) {
	out, err := c.exec.Run(ctx, "kubectl", "config", "get-contexts", "-o", "name")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}
	return lines, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/kubectl -run TestListContextsParsesKubectlOutput -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/ports/kubernetes.go internal/adapters/kubectl/discovery.go internal/adapters/kubectl/discovery_test.go
git commit -m "feat: add kubectl discovery adapter"
```

---

### Task 4: Implementar merge de catálogo y ranking Smart

**Files:**
- Create: `internal/app/catalog/service.go`
- Create: `internal/app/catalog/merge.go`
- Create: `internal/app/catalog/ranking.go`
- Test: `internal/app/catalog/service_test.go`

**Step 1: Write the failing test**

```go
package catalog

import (
	"testing"
	"time"

	"port-forward-tui/internal/domain"
)

func TestSmartRankingPrioritizesFavoriteRecentConfiguredTargets(t *testing.T) {
	now := time.Now()
	targets := []domain.Target{
		{Name: "worker", Type: domain.TargetTypePod},
		{Name: "admin", Type: domain.TargetTypeService, Favorite: true, PreferredLocalPort: 3001},
		{Name: "redis", Type: domain.TargetTypePod, LastUsedAt: now.Add(-1 * time.Hour)},
	}

	ranked := RankSmart(targets, now, "")

	if ranked[0].Name != "admin" {
		t.Fatalf("expected favorite configured target first, got %s", ranked[0].Name)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/catalog -run TestSmartRankingPrioritizesFavoriteRecentConfiguredTargets -v`
Expected: FAIL because `RankSmart` y campos de recencia no existen.

**Step 3: Write minimal implementation**

```go
package catalog

import (
	"sort"
	"strings"
	"time"

	"port-forward-tui/internal/domain"
)

func RankSmart(targets []domain.Target, now time.Time, query string) []domain.Target {
	type scored struct {
		target domain.Target
		score  int
	}

	items := make([]scored, 0, len(targets))
	for _, target := range targets {
		score := 0
		if target.Favorite {
			score += 100
		}
		if target.PreferredLocalPort != 0 {
			score += 25
		}
		if !target.LastUsedAt.IsZero() && now.Sub(target.LastUsedAt) < 24*time.Hour {
			score += 50
		}
		if query != "" && strings.Contains(strings.ToLower(target.Name), strings.ToLower(query)) {
			score += 75
		}
		items = append(items, scored{target: target, score: score})
	}

	sort.SliceStable(items, func(i, j int) bool { return items[i].score > items[j].score })

	result := make([]domain.Target, 0, len(items))
	for _, item := range items {
		result = append(result, item.target)
	}
	return result
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/catalog -run TestSmartRankingPrioritizesFavoriteRecentConfiguredTargets -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/app/catalog/service.go internal/app/catalog/merge.go internal/app/catalog/ranking.go internal/app/catalog/service_test.go internal/domain/target.go
git commit -m "feat: add target catalog merge and smart ranking"
```

---

### Task 5: Implementar persistencia local para favoritos, recientes y puertos preferidos

**Files:**
- Create: `internal/ports/config_store.go`
- Create: `internal/adapters/configfile/store.go`
- Test: `internal/adapters/configfile/store_test.go`

**Step 1: Write the failing test**

```go
package configfile

import (
	"testing"

	"port-forward-tui/internal/domain"
)

func TestStoreRoundTripPersistsTargetConfig(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	config := domain.AppConfig{
		Targets: map[string]domain.TargetConfig{
			"service:cco:admin": {Alias: "admin", PreferredLocalPort: 3001, Favorite: true},
		},
	}

	if err := store.Save(config); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Targets["service:cco:admin"].Alias != "admin" {
		t.Fatalf("expected alias persisted")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/configfile -run TestStoreRoundTripPersistsTargetConfig -v`
Expected: FAIL because `NewStore`, `AppConfig` o persistencia no existen.

**Step 3: Write minimal implementation**

```go
package configfile

import (
	"encoding/json"
	"os"
	"path/filepath"

	"port-forward-tui/internal/domain"
)

type Store struct{ path string }

func NewStore(baseDir string) Store {
	return Store{path: filepath.Join(baseDir, "config.json")}
}

func (s Store) Save(cfg domain.AppConfig) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s Store) Load() (domain.AppConfig, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return domain.AppConfig{Targets: map[string]domain.TargetConfig{}}, nil
	}
	if err != nil {
		return domain.AppConfig{}, err
	}
	var cfg domain.AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return domain.AppConfig{}, err
	}
	if cfg.Targets == nil {
		cfg.Targets = map[string]domain.TargetConfig{}
	}
	return cfg, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/configfile -run TestStoreRoundTripPersistsTargetConfig -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/ports/config_store.go internal/adapters/configfile/store.go internal/adapters/configfile/store_test.go internal/domain/config.go
git commit -m "feat: add local config persistence"
```

---

### Task 6: Implementar validación de puertos y runtime de port-forward

**Files:**
- Create: `internal/ports/runtime.go`
- Create: `internal/app/runtime/service.go`
- Create: `internal/adapters/kubectl/runtime.go`
- Test: `internal/app/runtime/service_test.go`

**Step 1: Write the failing test**

```go
package runtime

import (
	"context"
	"testing"

	"port-forward-tui/internal/domain"
)

func TestStartRejectsConflictingLocalPorts(t *testing.T) {
	runner := fakeRunner{}
	svc := NewService(runner)

	selection := []domain.ForwardRequest{
		{TargetID: "svc:admin", LocalPort: 3001, RemotePort: 3000},
		{TargetID: "pod:redis", LocalPort: 3001, RemotePort: 6379},
	}

	err := svc.StartMany(context.Background(), selection)
	if err == nil {
		t.Fatalf("expected conflict error")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/runtime -run TestStartRejectsConflictingLocalPorts -v`
Expected: FAIL because `NewService`, `ForwardRequest` y validación no existen.

**Step 3: Write minimal implementation**

```go
package runtime

import (
	"context"
	"fmt"

	"port-forward-tui/internal/domain"
)

type Runner interface {
	Start(context.Context, domain.ForwardRequest) (string, error)
}

type Service struct{ runner Runner }

func NewService(runner Runner) Service { return Service{runner: runner} }

func (s Service) StartMany(ctx context.Context, requests []domain.ForwardRequest) error {
	seen := map[int]struct{}{}
	for _, req := range requests {
		if _, exists := seen[req.LocalPort]; exists {
			return fmt.Errorf("local port %d already selected", req.LocalPort)
		}
		seen[req.LocalPort] = struct{}{}
	}
	for _, req := range requests {
		if _, err := s.runner.Start(ctx, req); err != nil {
			return err
		}
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/runtime -run TestStartRejectsConflictingLocalPorts -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/ports/runtime.go internal/app/runtime/service.go internal/adapters/kubectl/runtime.go internal/app/runtime/service_test.go internal/domain/forward_session.go
git commit -m "feat: add forward runtime service"
```

---

### Task 7: Crear layout TUI del workspace ligero

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/update.go`
- Modify: `internal/tui/view.go`
- Create: `internal/tui/keymap.go`
- Create: `internal/tui/state.go`
- Create: `internal/tui/components/header.go`
- Create: `internal/tui/components/catalog.go`
- Create: `internal/tui/components/selected_tab.go`
- Create: `internal/tui/components/running_tab.go`
- Test: `internal/tui/update_test.go`

**Step 1: Write the failing test**

```go
package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTabKeySwitchesFromSelectedToRunning(t *testing.T) {
	m := NewModel(Dependencies{})
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated := next.(Model)

	if updated.activeTab != TabRunning {
		t.Fatalf("expected running tab, got %q", updated.activeTab)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestTabKeySwitchesFromSelectedToRunning -v`
Expected: FAIL because no key handling exists.

**Step 3: Write minimal implementation**

```go
package tui

import tea "github.com/charmbracelet/bubbletea"

const TabRunning Tab = "running"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyTab {
			if m.activeTab == TabSelected {
				m.activeTab = TabRunning
			} else {
				m.activeTab = TabSelected
			}
		}
	}
	return m, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui -run TestTabKeySwitchesFromSelectedToRunning -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/tui/model.go internal/tui/update.go internal/tui/view.go internal/tui/keymap.go internal/tui/state.go internal/tui/components/header.go internal/tui/components/catalog.go internal/tui/components/selected_tab.go internal/tui/components/running_tab.go internal/tui/update_test.go
git commit -m "feat: add workspace tui layout"
```

---

### Task 8: Integrar catálogo, selección y edición de puertos en la tab Selected

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/update.go`
- Modify: `internal/tui/components/catalog.go`
- Modify: `internal/tui/components/selected_tab.go`
- Test: `internal/tui/selected_tab_test.go`

**Step 1: Write the failing test**

```go
package tui

import "testing"

func TestSelectingTargetAddsItToSelectedWithPreferredPort(t *testing.T) {
	m := NewModel(Dependencies{})
	m.catalog = []CatalogItem{{ID: "service:cco:admin", Label: "admin", PreferredLocalPort: 3001, RemotePort: 3000}}

	m.selectCurrentItem()

	if len(m.selected) != 1 || m.selected[0].LocalPort != 3001 {
		t.Fatalf("expected selected item with preferred local port, got %#v", m.selected)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestSelectingTargetAddsItToSelectedWithPreferredPort -v`
Expected: FAIL because selection state and helper do not exist.

**Step 3: Write minimal implementation**

```go
func (m *Model) selectCurrentItem() {
	if len(m.catalog) == 0 {
		return
	}
	item := m.catalog[m.cursor]
	m.selected = append(m.selected, SelectedItem{
		TargetID:   item.ID,
		Label:      item.Label,
		LocalPort:  item.PreferredLocalPort,
		RemotePort: item.RemotePort,
	})
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui -run TestSelectingTargetAddsItToSelectedWithPreferredPort -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/tui/model.go internal/tui/update.go internal/tui/components/catalog.go internal/tui/components/selected_tab.go internal/tui/selected_tab_test.go
git commit -m "feat: add target selection and port editing"
```

---

### Task 9: Integrar runtime y eventos en la tab Running

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/update.go`
- Modify: `internal/tui/components/running_tab.go`
- Test: `internal/tui/running_tab_test.go`
- Test: `test/integration/runtime_service_test.go`

**Step 1: Write the failing test**

```go
package tui

import "testing"

func TestRuntimeEventMarksForwardAsFailed(t *testing.T) {
	m := NewModel(Dependencies{})
	m.running = []RunningItem{{TargetID: "service:cco:admin", Status: StatusStarting}}

	next, _ := m.Update(RuntimeEvent{TargetID: "service:cco:admin", Status: StatusFailed, Err: "port in use"})
	updated := next.(Model)

	if updated.running[0].Status != StatusFailed || updated.running[0].Err != "port in use" {
		t.Fatalf("expected failed runtime state, got %#v", updated.running[0])
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestRuntimeEventMarksForwardAsFailed -v`
Expected: FAIL because runtime event handling does not exist.

**Step 3: Write minimal implementation**

```go
type RuntimeEvent struct {
	TargetID string
	Status   ForwardStatus
	Err      string
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case RuntimeEvent:
		for i := range m.running {
			if m.running[i].TargetID == msg.TargetID {
				m.running[i].Status = msg.Status
				m.running[i].Err = msg.Err
			}
		}
		return m, nil
	}
	// keep previous key handling
	return m, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui -run TestRuntimeEventMarksForwardAsFailed -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/tui/model.go internal/tui/update.go internal/tui/components/running_tab.go internal/tui/running_tab_test.go test/integration/runtime_service_test.go
git commit -m "feat: integrate runtime events into running tab"
```

---

### Task 10: Refinar UX operativa, wiring real y cobertura mínima

**Files:**
- Modify: `cmd/portfwd-tui/main.go`
- Modify: `internal/adapters/kubectl/discovery.go`
- Modify: `internal/adapters/kubectl/runtime.go`
- Modify: `internal/adapters/configfile/store.go`
- Modify: `internal/tui/view.go`
- Create: `README.md`
- Test: `test/integration/kubectl_discovery_test.go`

**Step 1: Write the failing test**

```go
package integration

import "testing"

func TestCatalogFlowLoadsConfiguredTargetsBeforeRendering(t *testing.T) {
	t.Skip("enable once composition root is wired")
}
```

**Step 2: Run test to verify it fails or is pending intentionally**

Run: `go test ./... -short`
Expected: unit tests PASS, integration pending or FAILING only where wiring is still incomplete.

**Step 3: Write minimal implementation**

```go
func main() {
	deps := bootstrapDependencies()
	p := tea.NewProgram(tui.NewModel(deps), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
```

Además en esta tarea:
- cablear dependencias reales en `main.go`
- asegurar cleanup ordenado con cancelación y stop de forwards activos
- documentar ubicación de config y atajos de teclado en `README.md`
- revisar textos de error para que sean accionables

**Step 4: Run tests to verify overall health**

Run: `go test ./...`
Expected: PASS en unit tests y tests de integración habilitados en entorno controlado.

**Step 5: Commit**

```bash
git add cmd/portfwd-tui/main.go internal/adapters/kubectl/discovery.go internal/adapters/kubectl/runtime.go internal/adapters/configfile/store.go internal/tui/view.go README.md test/integration/kubectl_discovery_test.go
git commit -m "feat: wire production dependencies and finalize tui mvp"
```

---

## Testing Strategy Summary

- **Unit tests first** para dominio, ranking, merge, validación de puertos, runtime state transitions.
- **Bubble Tea tests** centrados en `Update()` y estado, no snapshots masivos salvo componentes clave.
- **Integration tests** solo para adaptadores `kubectl` y wiring principal, idealmente detrás de `-short` o fixtures.
- **No build step as part of the change loop**; priorizar `go test ./...`.

## Delivery Notes

- Mantener el MVP alineado al script actual antes de pulir V1.1.
- Evitar introducir presets, session restore o health checks en esta fase.
- Si aparece complejidad extra en Bubble Tea, mover más lógica a `internal/app/...`, NO a `view.go`.
- Si el layout crece demasiado, dividir componentes pero mantener un solo `Model` raíz.

## Open Decisions To Resolve During Implementation

- Definir nombre final del módulo Go (`go mod init ...`).
- Confirmar formato final de persistencia (`json` vs `yaml`); JSON es suficiente para V1.
- Decidir si la detección de puertos ocupados del host será obligatoria en V1 o empezará solo con conflictos internos + activos en app.
