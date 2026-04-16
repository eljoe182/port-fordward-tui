package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"cco-port-forward-tui/internal/ports"
)

type Tab string

const (
	TabSelected Tab = "selected"
	TabRunning  Tab = "running"
)

type Dependencies struct {
	Discovery   ports.KubernetesDiscovery
	ConfigStore ports.ConfigStore
	Runtime     ports.ForwardRunner
}

type Model struct {
	deps      Dependencies
	activeTab Tab
	catalog   []CatalogItem
	cursor    int
	selected  []SelectedItem
	running   []RunningItem
	errMsg    string
}

func NewModel(deps Dependencies) Model {
	return Model{
		deps:      deps,
		activeTab: TabSelected,
		catalog:   []CatalogItem{},
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m *Model) selectCurrentItem() {
	if len(m.catalog) == 0 {
		return
	}
	if m.cursor < 0 || m.cursor >= len(m.catalog) {
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
