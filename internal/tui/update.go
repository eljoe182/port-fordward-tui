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
		return m, m.applyForwardStarted(msg)
	case forwardFailedMsg:
		m.applyForwardFailed(msg)
		return m, nil
	case forwardBatchMsg:
		var cmds []tea.Cmd
		for _, started := range msg.started {
			if cmd := m.applyForwardStarted(started); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		for _, failed := range msg.failed {
			m.applyForwardFailed(failed)
		}
		if len(cmds) == 0 {
			return m, nil
		}
		return m, tea.Batch(cmds...)
	case forwardStoppedMsg:
		if msg.RemoveSelected && m.pendingRemovalMatches(msg.Ref, msg.SessionID) {
			m.completeSelectedRemoval(msg.Ref)
			return m, nil
		}
		removedIdx := runningIndexBySession(m.running, msg.SessionID)
		m.running = removeRunningBySession(m.running, msg.SessionID)
		m.runningCursor = cursorAfterRemoval(m.runningCursor, removedIdx, len(m.running))
		return m, nil
	case forwardStopFailedMsg:
		errMsg := actionableError(msg.Err)
		if idx := runningIndexBySession(m.running, msg.SessionID); idx >= 0 {
			m.running[idx].Err = errMsg
		}
		if msg.RemoveSelected && m.pendingRemovalMatches(msg.Ref, msg.SessionID) {
			delete(m.pendingRemovals, msg.Ref)
			m.errMsg = errMsg
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
		if m.activeTab == TabSelected {
			return m.removeSelectedUnderCursor()
		}
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

func (m Model) removeSelectedUnderCursor() (tea.Model, tea.Cmd) {
	if m.activeTab != TabSelected || m.selectedCursor < 0 || m.selectedCursor >= len(m.selected) {
		return m, nil
	}

	ref := m.selected[m.selectedCursor].ref()
	m.errMsg = ""
	runningIdx := runningIndexByRef(m.running, ref)
	if runningIdx < 0 {
		m.completeSelectedRemoval(ref)
		return m, nil
	}
	if m.hasPendingRemoval(ref) {
		return m, nil
	}

	running := m.running[runningIdx]
	if running.Status == StatusFailed || running.Status == StatusStopped {
		m.completeSelectedRemoval(ref)
		return m, nil
	}
	if m.pendingRemovals == nil {
		m.pendingRemovals = map[forwardRef]string{}
	}
	if running.Status == StatusStarting {
		m.pendingRemovals[ref] = ""
		return m, nil
	}
	if running.SessionID == "" {
		m.errMsg = "forward session unavailable"
		return m, nil
	}
	if m.deps.Runtime == nil {
		m.errMsg = "forward runtime unavailable"
		return m, nil
	}
	m.pendingRemovals[ref] = running.SessionID
	return m, stopForwardCmd(m.ctx, m.deps.Runtime, running.SessionID, ref, true)
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
	if item.Status == StatusStarting || item.SessionID == "" {
		m.errMsg = "forward is still starting"
		return m, nil
	}
	m.errMsg = ""
	return m, stopForwardCmd(m.ctx, m.deps.Runtime, item.SessionID, item.ref(), false)
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
	m.running[idx].SessionID = ""
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

func (m *Model) applyForwardStarted(msg forwardStartedMsg) tea.Cmd {
	idx := runningIndexByRef(m.running, msg.Ref)
	if idx < 0 {
		return nil
	}
	m.running[idx].Status = StatusRunning
	m.running[idx].SessionID = msg.SessionID
	m.running[idx].Err = ""
	if event, ok := m.pendingEvents[msg.SessionID]; ok {
		delete(m.pendingEvents, msg.SessionID)
		m.applyRuntimeEvent(event)
	}
	return m.continuePendingRemoval(msg.Ref)
}

func (m *Model) applyForwardFailed(msg forwardFailedMsg) {
	idx := runningIndexByRef(m.running, msg.Ref)
	if idx < 0 {
		return
	}
	m.running[idx].Status = StatusFailed
	m.running[idx].Err = actionableError(msg.Err)
	if m.hasPendingRemoval(msg.Ref) {
		m.completeSelectedRemoval(msg.Ref)
	}
}

func (m *Model) applyRuntimeEvent(event domain.ForwardEvent) bool {
	idx := runningIndexBySession(m.running, event.SessionID)
	if idx < 0 {
		return false
	}
	ref := m.running[idx].ref()

	switch event.Status {
	case domain.ForwardStatusStopped:
		if m.hasPendingRemoval(ref) {
			m.completeSelectedRemoval(ref)
			return true
		}
		removedIdx := idx
		m.running = removeRunningBySession(m.running, event.SessionID)
		m.runningCursor = cursorAfterRemoval(m.runningCursor, removedIdx, len(m.running))
	case domain.ForwardStatusFailed:
		if m.hasPendingRemoval(ref) {
			m.completeSelectedRemoval(ref)
			return true
		}
		m.running[idx].Status = StatusFailed
		m.running[idx].Err = actionableError(event.Err)
	case domain.ForwardStatusRunning:
		m.running[idx].Status = StatusRunning
		m.running[idx].Err = ""
	}
	return true
}

func (m *Model) continuePendingRemoval(ref forwardRef) tea.Cmd {
	if !m.hasPendingRemoval(ref) {
		return nil
	}
	idx := runningIndexByRef(m.running, ref)
	if idx < 0 {
		m.completeSelectedRemoval(ref)
		return nil
	}
	item := m.running[idx]
	if item.Status == StatusFailed || item.Status == StatusStopped {
		m.completeSelectedRemoval(ref)
		return nil
	}
	if item.Status == StatusStarting || item.SessionID == "" {
		return nil
	}
	if m.deps.Runtime == nil {
		delete(m.pendingRemovals, ref)
		m.errMsg = "forward runtime unavailable"
		return nil
	}
	m.pendingRemovals[ref] = item.SessionID
	return stopForwardCmd(m.ctx, m.deps.Runtime, item.SessionID, ref, true)
}

func (m Model) hasPendingRemoval(ref forwardRef) bool {
	_, ok := m.pendingRemovals[ref]
	return ok
}

func (m Model) pendingRemovalMatches(ref forwardRef, sessionID string) bool {
	pendingSessionID, ok := m.pendingRemovals[ref]
	return ok && pendingSessionID == sessionID
}

func (m *Model) completeSelectedRemoval(ref forwardRef) {
	selectedIdx := selectedIndexByRef(m.selected, ref)
	runningIdx := runningIndexByRef(m.running, ref)
	m.selected = removeSelectedByRef(m.selected, ref)
	m.running = removeRunningByRef(m.running, ref)
	delete(m.pendingRemovals, ref)
	m.selectedCursor = cursorAfterRemoval(m.selectedCursor, selectedIdx, len(m.selected))
	m.runningCursor = cursorAfterRemoval(m.runningCursor, runningIdx, len(m.running))
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

func removeSelectedByRef(selected []SelectedItem, ref forwardRef) []SelectedItem {
	out := selected[:0]
	for _, item := range selected {
		if item.ref() != ref {
			out = append(out, item)
		}
	}
	return out
}

func selectedIndexByRef(selected []SelectedItem, ref forwardRef) int {
	for i, item := range selected {
		if item.ref() == ref {
			return i
		}
	}
	return -1
}

func removeRunningByRef(running []RunningItem, ref forwardRef) []RunningItem {
	out := running[:0]
	for _, item := range running {
		if item.ref() != ref {
			out = append(out, item)
		}
	}
	return out
}

func cursorAfterRemoval(cursor, removedIdx, remaining int) int {
	if remaining == 0 {
		return 0
	}
	if removedIdx >= 0 && removedIdx < cursor {
		cursor--
	}
	if cursor < 0 {
		return 0
	}
	if cursor >= remaining {
		return remaining - 1
	}
	return cursor
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
