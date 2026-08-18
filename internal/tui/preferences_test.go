package tui

import (
	"testing"

	"port-forward-tui/internal/domain"
)

func TestPersistRecentSelectionsDoesNotPersistAutomaticPort(t *testing.T) {
	store := &fakeStore{cfg: domain.AppConfig{
		CurrentContext:   "prod",
		CurrentNamespace: "default",
		Targets:          map[string]domain.TargetConfig{},
	}}
	m := NewModel(Dependencies{ConfigStore: store})
	m.contextName = "dev"
	m.namespace = "cco"
	item := SelectedItem{
		Context: "dev", Namespace: "cco", Type: "service", TargetID: "service:cco:admin", Label: "admin", LocalPort: 3003, RemotePort: 3000,
	}

	cmd := m.persistRecentSelections([]SelectedItem{item})
	if msg := cmd(); msg != nil {
		t.Fatalf("expected successful persistence, got %T", msg)
	}

	entry := store.cfg.Targets[item.TargetID]
	if entry.PreferredLocalPort != 0 {
		t.Fatalf("expected automatic port to remain transient, got %+v", entry)
	}
	if entry.LastUsedAt.IsZero() {
		t.Fatalf("expected recency metadata to be persisted")
	}
	if store.cfg.CurrentContext != "prod" || store.cfg.CurrentNamespace != "default" {
		t.Fatalf("expected browsing scope unchanged, got %s/%s", store.cfg.CurrentContext, store.cfg.CurrentNamespace)
	}
}

func TestPersistRecentSelectionsPreservesExistingPreferredPort(t *testing.T) {
	const targetID = "service:cco:admin"
	store := &fakeStore{cfg: domain.AppConfig{Targets: map[string]domain.TargetConfig{
		targetID: {Type: domain.TargetTypeService, Namespace: "cco", Name: "admin", PreferredLocalPort: 3000},
	}}}
	m := NewModel(Dependencies{ConfigStore: store})
	item := SelectedItem{
		Context: "prod", Namespace: "cco", Type: "service", TargetID: targetID, Label: "admin", LocalPort: 3003, RemotePort: 3000,
	}

	cmd := m.persistRecentSelections([]SelectedItem{item})
	if msg := cmd(); msg != nil {
		t.Fatalf("expected successful persistence, got %T", msg)
	}

	if got := store.cfg.Targets[targetID].PreferredLocalPort; got != 3000 {
		t.Fatalf("expected preferred port 3000 preserved, got %d", got)
	}
}
