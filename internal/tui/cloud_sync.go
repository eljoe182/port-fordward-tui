package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"port-forward-tui/internal/ports"
)

type cloudSyncResultMsg struct {
	result ports.CredentialSyncResult
	err    error
}

func (m Model) startCloudCredentialSync() (tea.Model, tea.Cmd) {
	if m.deps.CloudSync == nil {
		m.errMsg = "Azure AKS sync is unavailable"
		return m, nil
	}
	if m.syncing {
		return m, nil
	}
	m.syncing = true
	m.errMsg = "Syncing Azure AKS credentials…"
	m.pendingHeader = ""
	return m, syncCloudCredentialsCmd(m.ctx, m.deps.CloudSync)
}

func syncCloudCredentialsCmd(ctx context.Context, syncer ports.CloudCredentialSyncer) tea.Cmd {
	return func() tea.Msg {
		result, err := syncer.Sync(ctx)
		return cloudSyncResultMsg{result: result, err: err}
	}
}

func (m Model) applyCloudSyncResult(msg cloudSyncResultMsg) (tea.Model, tea.Cmd) {
	m.syncing = false
	if msg.err != nil {
		m.pendingHeader = ""
		m.errMsg = msg.err.Error()
		return m, nil
	}

	summary := formatCloudSyncSummary(msg.result)
	m.pendingHeader = summary
	m.errMsg = summary
	return m, loadCatalogCmd(m.ctx, m.deps, m.loadOptions())
}

func formatCloudSyncSummary(result ports.CredentialSyncResult) string {
	provider := result.Provider
	if provider == "" {
		provider = "cloud"
	}
	if len(result.Failed) == 0 {
		return fmt.Sprintf("Synced %d %s cluster(s)", result.Synced, provider)
	}
	names := make([]string, 0, len(result.Failed))
	for _, f := range result.Failed {
		names = append(names, f.Name)
	}
	return fmt.Sprintf(
		"Synced %d %s cluster(s); %d failed (%s)",
		result.Synced,
		provider,
		len(result.Failed),
		strings.Join(names, ", "),
	)
}
