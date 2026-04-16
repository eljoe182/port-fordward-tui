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
	case forwardStartedMsg:
		for i := range m.running {
			if m.running[i].TargetID == msg.TargetID {
				m.running[i].Status = StatusRunning
				m.running[i].SessionID = msg.SessionID
				m.running[i].Err = ""
			}
		}
		return m, nil
	case forwardFailedMsg:
		for i := range m.running {
			if m.running[i].TargetID == msg.TargetID {
				m.running[i].Status = StatusFailed
				m.running[i].Err = msg.Err
			}
		}
		return m, nil
	case forwardStoppedMsg:
		m.running = removeRunningByTarget(m.running, msg.TargetID)
		if m.runningCursor >= len(m.running) {
			m.runningCursor = maxInt(0, len(m.running)-1)
		}
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyTab:
		if m.activeTab == TabSelected {
			m.activeTab = TabRunning
		} else {
			m.activeTab = TabSelected
		}
		return m, nil
	case tea.KeyEnter:
		m.selectCurrentItem()
		return m, nil
	case tea.KeyUp:
		m.moveCursor(-1)
		return m, nil
	case tea.KeyDown:
		m.moveCursor(1)
		return m, nil
	case tea.KeyEsc:
		m.errMsg = ""
		return m, nil
	}

	switch string(msg.Runes) {
	case "q":
		return m, tea.Quit
	case "j":
		m.moveCursor(1)
		return m, nil
	case "k":
		m.moveCursor(-1)
		return m, nil
	case "s":
		return m.startSelectedForwards()
	case "x":
		return m.stopRunningUnderCursor()
	}
	return m, nil
}

func (m Model) startSelectedForwards() (tea.Model, tea.Cmd) {
	if len(m.selected) == 0 || m.deps.Runtime == nil {
		return m, nil
	}

	next := make([]RunningItem, 0, len(m.selected))
	for _, item := range m.selected {
		if isAlreadyRunning(m.running, item.TargetID) {
			continue
		}
		next = append(next, RunningItem{
			TargetID:   item.TargetID,
			Label:      item.Label,
			LocalPort:  item.LocalPort,
			RemotePort: item.RemotePort,
			Status:     StatusStarting,
		})
	}
	m.running = append(m.running, next...)
	m.activeTab = TabRunning

	cmds := startForwardsCmds(m.ctx, m.deps.Runtime, m.selected, m.contextName, m.namespace)
	return m, tea.Batch(cmds...)
}

func (m Model) stopRunningUnderCursor() (tea.Model, tea.Cmd) {
	if m.activeTab != TabRunning || len(m.running) == 0 || m.deps.Runtime == nil {
		return m, nil
	}
	idx := m.runningCursor
	if idx < 0 || idx >= len(m.running) {
		return m, nil
	}
	item := m.running[idx]
	m.running = removeRunningByTarget(m.running, item.TargetID)
	if m.runningCursor >= len(m.running) {
		m.runningCursor = maxInt(0, len(m.running)-1)
	}
	return m, stopForwardCmd(m.ctx, m.deps.Runtime, item.TargetID, item.SessionID)
}

func isAlreadyRunning(running []RunningItem, targetID string) bool {
	for _, r := range running {
		if r.TargetID == targetID {
			return true
		}
	}
	return false
}

func removeRunningByTarget(running []RunningItem, targetID string) []RunningItem {
	out := running[:0]
	for _, r := range running {
		if r.TargetID != targetID {
			out = append(out, r)
		}
	}
	return out
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
