package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if isTabKey(msg) {
			if m.activeTab == TabSelected {
				m.activeTab = TabRunning
			} else {
				m.activeTab = TabSelected
			}
		}
	}
	return m, nil
}
