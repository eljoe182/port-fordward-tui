package tui

import tea "github.com/charmbracelet/bubbletea"

type Tab string

const (
	TabSelected Tab = "selected"
	TabRunning  Tab = "running"
)

type Dependencies struct{}

type Model struct {
	activeTab Tab
	catalog   []string
}

func NewModel(_ Dependencies) Model {
	return Model{activeTab: TabSelected, catalog: []string{}}
}

func (m Model) Init() tea.Cmd { return nil }
