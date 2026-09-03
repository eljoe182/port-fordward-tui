package tui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"port-forward-tui/internal/ports"
)

type fakeCloudSyncer struct {
	provider string
	result   ports.CredentialSyncResult
	err      error
	calls    int
}

func (f *fakeCloudSyncer) Provider() string {
	if f.provider != "" {
		return f.provider
	}
	return "Azure AKS"
}

func (f *fakeCloudSyncer) Sync(context.Context) (ports.CredentialSyncResult, error) {
	f.calls++
	if f.err != nil {
		return ports.CredentialSyncResult{Provider: f.Provider()}, f.err
	}
	result := f.result
	if result.Provider == "" {
		result.Provider = f.Provider()
	}
	return result, nil
}

func TestCtrlRWithoutCloudSyncShowsError(t *testing.T) {
	m := NewModel(Dependencies{})
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	updated := next.(Model)
	if cmd != nil {
		t.Fatal("expected no command when CloudSync is nil")
	}
	if updated.errMsg != "Azure AKS sync is unavailable" {
		t.Fatalf("unexpected errMsg: %q", updated.errMsg)
	}
	if updated.syncing {
		t.Fatal("expected syncing=false")
	}
}

func TestCtrlRStartsSyncAndBlocksReentry(t *testing.T) {
	syncer := &fakeCloudSyncer{
		result: ports.CredentialSyncResult{Synced: 2},
	}
	m := NewModel(Dependencies{CloudSync: syncer})

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	updated := next.(Model)
	if !updated.syncing {
		t.Fatal("expected syncing=true")
	}
	if updated.errMsg != "Syncing Azure AKS credentials…" {
		t.Fatalf("unexpected errMsg: %q", updated.errMsg)
	}
	if cmd == nil {
		t.Fatal("expected sync command")
	}

	next, cmd2 := updated.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	blocked := next.(Model)
	if cmd2 != nil {
		t.Fatal("expected no command while already syncing")
	}
	if !blocked.syncing {
		t.Fatal("expected syncing to remain true")
	}
	if syncer.calls != 0 {
		t.Fatalf("sync should not run until cmd executes, calls=%d", syncer.calls)
	}
}

func TestCloudSyncResultSuccessReloadsCatalog(t *testing.T) {
	syncer := &fakeCloudSyncer{
		result: ports.CredentialSyncResult{Provider: "Azure AKS", Synced: 3},
	}
	m := NewModel(Dependencies{CloudSync: syncer})
	m.syncing = true

	next, cmd := m.Update(cloudSyncResultMsg{
		result: ports.CredentialSyncResult{Provider: "Azure AKS", Synced: 3},
	})
	updated := next.(Model)
	if updated.syncing {
		t.Fatal("expected syncing=false after result")
	}
	if updated.errMsg != "Synced 3 Azure AKS cluster(s)" {
		t.Fatalf("unexpected errMsg: %q", updated.errMsg)
	}
	if updated.pendingHeader != "Synced 3 Azure AKS cluster(s)" {
		t.Fatalf("unexpected pendingHeader: %q", updated.pendingHeader)
	}
	if cmd == nil {
		t.Fatal("expected catalog reload command")
	}
}

func TestCloudSyncResultHardErrorSkipsReload(t *testing.T) {
	m := NewModel(Dependencies{CloudSync: &fakeCloudSyncer{}})
	m.syncing = true

	next, cmd := m.Update(cloudSyncResultMsg{err: errors.New("az: not logged in")})
	updated := next.(Model)
	if updated.syncing {
		t.Fatal("expected syncing=false")
	}
	if updated.errMsg != "az: not logged in" {
		t.Fatalf("unexpected errMsg: %q", updated.errMsg)
	}
	if updated.pendingHeader != "" {
		t.Fatalf("expected empty pendingHeader, got %q", updated.pendingHeader)
	}
	if cmd != nil {
		t.Fatal("expected no catalog reload on hard error")
	}
}

func TestCatalogLoadedPreservesPendingHeader(t *testing.T) {
	m := NewModel(Dependencies{})
	m.pendingHeader = "Synced 1 Azure AKS cluster(s)"
	m.errMsg = "old"

	next, _ := m.Update(catalogLoadedMsg{result: CatalogResult{
		Contexts: []string{"dev"},
		Context:  "dev",
	}})
	updated := next.(Model)
	if updated.errMsg != "Synced 1 Azure AKS cluster(s)" {
		t.Fatalf("expected pending header preserved, got %q", updated.errMsg)
	}
	if updated.pendingHeader != "" {
		t.Fatalf("expected pendingHeader cleared, got %q", updated.pendingHeader)
	}
}

func TestFormatCloudSyncSummaryPartialFailures(t *testing.T) {
	got := formatCloudSyncSummary(ports.CredentialSyncResult{
		Provider: "Azure AKS",
		Synced:   1,
		Failed: []ports.CredentialSyncFailure{
			{Name: "rg/bad"},
		},
	})
	want := "Synced 1 Azure AKS cluster(s); 1 failed (rg/bad)"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
