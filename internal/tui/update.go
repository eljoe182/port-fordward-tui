package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case catalogLoadedMsg:
		m.contextName = msg.result.Context
		m.namespace = msg.result.Namespace
		m.catalog = msg.result.Items
		if m.cursor >= len(m.catalog) {
			m.cursor = 0
		}
		m.errMsg = ""
		return m, nil
	case catalogErrorMsg:
		m.errMsg = msg.err.Error()
		return m, nil
	case RuntimeEvent:
		for i := range m.running {
			if m.running[i].TargetID == msg.TargetID {
				m.running[i].Status = msg.Status
				m.running[i].Err = msg.Err
			}
		}
		return m, nil
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
