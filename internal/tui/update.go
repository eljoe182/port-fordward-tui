package tui

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"

	"port-forward-tui/internal/app/catalog"
	"port-forward-tui/internal/domain"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case catalogLoadedMsg:
		m.contexts = append([]string(nil), msg.result.Contexts...)
		m.namespaces = append([]string(nil), msg.result.Namespaces...)
		m.contextName = msg.result.Context
		m.namespace = msg.result.Namespace
		m.query = msg.result.Query
		m.filterMode = catalog.FilterMode(msg.result.Filter)
		m.sortMode = catalog.SortMode(msg.result.Sort)
		m.catalog = msg.result.Items
		if m.cursor >= len(m.catalog) {
			m.cursor = 0
		}
		m.errMsg = ""
		return m, nil
	case catalogErrorMsg:
		m.errMsg = msg.err.Error()
		return m, nil
	case forwardStartedMsg:
		m.applyForwardStarted(msg)
		return m, nil
	case forwardFailedMsg:
		m.applyForwardFailed(msg)
		return m, nil
	case forwardBatchMsg:
		for _, started := range msg.started {
			m.applyForwardStarted(started)
		}
		for _, failed := range msg.failed {
			m.applyForwardFailed(failed)
		}
		return m, nil
	case forwardStoppedMsg:
		m.running = removeRunningBySession(m.running, msg.SessionID)
		if m.runningCursor >= len(m.running) {
			m.runningCursor = maxInt(0, len(m.running)-1)
		}
		return m, nil
	case forwardStopFailedMsg:
		if idx := runningIndexBySession(m.running, msg.SessionID); idx >= 0 {
			m.running[idx].Err = actionableError(msg.Err)
		}
		return m, nil
	case forwardEventMsg:
		return m.applyForwardEvent(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.modalKind != ModalNone {
		return m.handleModalKey(msg)
	}
	if m.editingPort {
		return m.handleEditPortKey(msg)
	}

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
	case "/":
		return m.openSearchModal(), nil
	case "c":
		return m.openContextModal(), nil
	case "n":
		return m.openNamespaceModal(), nil
	case "t":
		return m.openFilterModal(), nil
	case "o":
		return m.openSortModal(), nil
	case "r":
		return m, m.refreshCatalog()
	case "q":
		return m, tea.Quit
	case "j":
		m.moveCursor(1)
		return m, nil
	case "k":
		m.moveCursor(-1)
		return m, nil
	case "J":
		return m.moveActiveCursor(1), nil
	case "K":
		return m.moveActiveCursor(-1), nil
	case "s":
		return m.startSelectedForwards()
	case "x":
		return m.stopRunningUnderCursor()
	case "R":
		return m.retryRunningUnderCursor()
	case "e":
		return m.enterPortEditMode(), nil
	case "f":
		if len(m.catalog) == 0 || m.cursor < 0 || m.cursor >= len(m.catalog) {
			return m, nil
		}
		m.catalog[m.cursor].Favorite = !m.catalog[m.cursor].Favorite
		return m, m.toggleFavoriteCurrentItem()
	}
	return m, nil
}

func (m Model) moveActiveCursor(delta int) Model {
	switch m.activeTab {
	case TabSelected:
		if len(m.selected) == 0 {
			m.selectedCursor = 0
			return m
		}
		next := m.selectedCursor + delta
		if next < 0 {
			next = 0
		}
		if next >= len(m.selected) {
			next = len(m.selected) - 1
		}
		m.selectedCursor = next
	case TabRunning:
		if len(m.running) == 0 {
			m.runningCursor = 0
			return m
		}
		next := m.runningCursor + delta
		if next < 0 {
			next = 0
		}
		if next >= len(m.running) {
			next = len(m.running) - 1
		}
		m.runningCursor = next
	default:
		m.moveCursor(delta)
	}
	return m
}

func (m Model) enterPortEditMode() Model {
	if m.activeTab != TabSelected || len(m.selected) == 0 {
		return m
	}
	if m.selectedCursor < 0 || m.selectedCursor >= len(m.selected) {
		return m
	}
	m.editingPort = true
	m.portBuffer = strconv.Itoa(m.selected[m.selectedCursor].LocalPort)
	m.errMsg = ""
	return m
}

func (m Model) handleEditPortKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.editingPort = false
		m.portBuffer = ""
		m.errMsg = ""
		return m, nil
	case tea.KeyEnter:
		port, err := strconv.Atoi(m.portBuffer)
		if err != nil || port < 1 || port > 65535 {
			m.errMsg = fmt.Sprintf("invalid port %q (must be 1..65535)", m.portBuffer)
			return m, nil
		}
		m.selected[m.selectedCursor].LocalPort = port
		selected := m.selected[m.selectedCursor]
		m.editingPort = false
		m.portBuffer = ""
		m.errMsg = ""
		return m, m.persistSelectedPort(selected)
	case tea.KeyBackspace:
		if len(m.portBuffer) > 0 {
			m.portBuffer = m.portBuffer[:len(m.portBuffer)-1]
		}
		return m, nil
	case tea.KeyCtrlC:
		return m, tea.Quit
	}
	for _, r := range msg.Runes {
		if r >= '0' && r <= '9' && len(m.portBuffer) < 5 {
			m.portBuffer += string(r)
		}
	}
	return m, nil
}

func (m Model) startSelectedForwards() (tea.Model, tea.Cmd) {
	if len(m.selected) == 0 || !m.deps.RuntimeApp.Available() {
		return m, nil
	}

	active := activeForwardSessions(m.running)
	next := make([]RunningItem, 0, len(m.selected))
	toStart := make([]SelectedItem, 0, len(m.selected))
	for _, item := range m.selected {
		if isAlreadyRunning(m.running, item.ref()) {
			continue
		}
		toStart = append(toStart, item)
		next = append(next, RunningItem{
			TargetID:   item.TargetID,
			Context:    item.Context,
			Namespace:  item.Namespace,
			Type:       item.Type,
			Label:      item.Label,
			LocalPort:  item.LocalPort,
			RemotePort: item.RemotePort,
			Status:     StatusStarting,
		})
	}
	if len(toStart) == 0 {
		return m, nil
	}
	m.running = append(m.running, next...)
	m.activeTab = TabRunning

	cmds := []tea.Cmd{startForwardsCmd(m.ctx, m.deps.RuntimeApp, toStart, active)}
	if persist := m.persistRecentSelections(toStart); persist != nil {
		cmds = append(cmds, persist)
	}
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
	if item.SessionID == "" {
		m.errMsg = "forward is still starting"
		return m, nil
	}
	m.errMsg = ""
	return m, stopForwardCmd(m.ctx, m.deps.Runtime, item.SessionID)
}

func (m Model) retryRunningUnderCursor() (tea.Model, tea.Cmd) {
	if m.activeTab != TabRunning || len(m.running) == 0 || !m.deps.RuntimeApp.Available() {
		return m, nil
	}
	idx := m.runningCursor
	if idx < 0 || idx >= len(m.running) {
		return m, nil
	}
	item := m.running[idx]
	if item.Status != StatusFailed && item.Status != StatusStopped {
		return m, nil
	}
	active := activeForwardSessions(m.running)
	m.running[idx].Status = StatusStarting
	m.running[idx].Err = ""
	return m, retryForwardCmd(m.ctx, m.deps.RuntimeApp, m.running[idx], active)
}

func (m Model) applyForwardEvent(msg forwardEventMsg) (tea.Model, tea.Cmd) {
	event := msg.event
	if !m.applyRuntimeEvent(event) && m.shouldBufferEvent(event) {
		if m.pendingEvents == nil {
			m.pendingEvents = map[string]domain.ForwardEvent{}
		}
		m.pendingEvents[event.SessionID] = event
	}
	return m, listenForwardEventsCmd(m.deps.Runtime)
}

func (m *Model) applyForwardStarted(msg forwardStartedMsg) {
	idx := runningIndexByRef(m.running, msg.Ref)
	if idx < 0 {
		return
	}
	m.running[idx].Status = StatusRunning
	m.running[idx].SessionID = msg.SessionID
	m.running[idx].Err = ""
	if event, ok := m.pendingEvents[msg.SessionID]; ok {
		delete(m.pendingEvents, msg.SessionID)
		m.applyRuntimeEvent(event)
	}
}

func (m *Model) applyForwardFailed(msg forwardFailedMsg) {
	idx := runningIndexByRef(m.running, msg.Ref)
	if idx < 0 {
		return
	}
	m.running[idx].Status = StatusFailed
	m.running[idx].Err = actionableError(msg.Err)
}

func (m *Model) applyRuntimeEvent(event domain.ForwardEvent) bool {
	idx := runningIndexBySession(m.running, event.SessionID)
	if idx < 0 {
		return false
	}

	switch event.Status {
	case domain.ForwardStatusStopped:
		m.running = removeRunningBySession(m.running, event.SessionID)
		if m.runningCursor >= len(m.running) {
			m.runningCursor = maxInt(0, len(m.running)-1)
		}
	case domain.ForwardStatusFailed:
		m.running[idx].Status = StatusFailed
		m.running[idx].Err = actionableError(event.Err)
	case domain.ForwardStatusRunning:
		m.running[idx].Status = StatusRunning
		m.running[idx].Err = ""
	}
	return true
}

func (m Model) shouldBufferEvent(event domain.ForwardEvent) bool {
	if event.SessionID == "" {
		return false
	}
	for _, item := range m.running {
		if item.SessionID == "" && item.TargetID == event.TargetID {
			return true
		}
	}
	return false
}

func isAlreadyRunning(running []RunningItem, ref forwardRef) bool {
	for _, r := range running {
		if r.ref() == ref {
			return true
		}
	}
	return false
}

func runningIndexByRef(running []RunningItem, ref forwardRef) int {
	for i, item := range running {
		if item.ref() == ref {
			return i
		}
	}
	return -1
}

func runningIndexBySession(running []RunningItem, sessionID string) int {
	if sessionID == "" {
		return -1
	}
	for i, item := range running {
		if item.SessionID == sessionID {
			return i
		}
	}
	return -1
}

func removeRunningBySession(running []RunningItem, sessionID string) []RunningItem {
	out := running[:0]
	for _, r := range running {
		if r.SessionID != sessionID {
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
