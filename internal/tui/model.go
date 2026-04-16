package tui

import (
	"context"

	appruntime "cco-port-forward-tui/internal/app/runtime"
	tea "github.com/charmbracelet/bubbletea"

	"cco-port-forward-tui/internal/app/catalog"
	"cco-port-forward-tui/internal/ports"
)

type Tab string
type ModalKind string

const (
	TabSelected Tab = "selected"
	TabRunning  Tab = "running"

	ModalNone      ModalKind = ""
	ModalContext   ModalKind = "context"
	ModalNamespace ModalKind = "namespace"
	ModalFilter    ModalKind = "filter"
	ModalSort      ModalKind = "sort"
	ModalSearch    ModalKind = "search"
)

type Dependencies struct {
	Discovery   ports.KubernetesDiscovery
	ConfigStore ports.ConfigStore
	Runtime     ports.ForwardRunner
	RuntimeApp  appruntime.Service
}

type Model struct {
	deps           Dependencies
	ctx            context.Context
	activeTab      Tab
	contexts       []string
	namespaces     []string
	contextName    string
	namespace      string
	query          string
	queryBuffer    string
	filterMode     catalog.FilterMode
	sortMode       catalog.SortMode
	modalKind      ModalKind
	modalCursor    int
	modalInput     string
	catalog        []CatalogItem
	cursor         int
	selected       []SelectedItem
	selectedCursor int
	editingPort    bool
	portBuffer     string
	running        []RunningItem
	runningCursor  int
	width          int
	height         int
	errMsg         string
}

func NewModel(deps Dependencies) Model {
	return Model{
		deps:       deps,
		ctx:        context.Background(),
		activeTab:  TabSelected,
		filterMode: catalog.FilterAll,
		sortMode:   catalog.SortSmart,
		catalog:    []CatalogItem{},
	}
}

func (m Model) WithContext(ctx context.Context) Model {
	m.ctx = ctx
	return m
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	if m.deps.Discovery != nil && m.deps.ConfigStore != nil {
		cmds = append(cmds, loadCatalogCmd(m.ctx, m.deps, m.loadOptions()))
	}
	if listen := listenForwardEventsCmd(m.deps.Runtime); listen != nil {
		cmds = append(cmds, listen)
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m Model) loadOptions() catalog.LoadOptions {
	return catalog.LoadOptions{Query: m.query, Filter: m.filterMode, Sort: m.sortMode}
}

func (m *Model) selectCurrentItem() {
	if len(m.catalog) == 0 {
		return
	}
	if m.cursor < 0 || m.cursor >= len(m.catalog) {
		return
	}
	item := m.catalog[m.cursor]
	for _, existing := range m.selected {
		if existing.TargetID == item.ID {
			return
		}
	}
	m.selected = append(m.selected, SelectedItem{
		TargetID:   item.ID,
		Label:      item.Label,
		LocalPort:  item.PreferredLocalPort,
		RemotePort: item.RemotePort,
	})
}

func (m *Model) moveCursor(delta int) {
	if len(m.catalog) == 0 {
		return
	}
	next := m.cursor + delta
	if next < 0 {
		next = 0
	}
	if next >= len(m.catalog) {
		next = len(m.catalog) - 1
	}
	m.cursor = next
}
