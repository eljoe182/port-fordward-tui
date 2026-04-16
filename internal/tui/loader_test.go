package tui

import (
	"context"
	"testing"

	"cco-port-forward-tui/internal/domain"
)

type fakeDiscovery struct {
	currentContext string
	namespaces     []string
	targets        []domain.Target
	targetsErr     error
}

func (f fakeDiscovery) CurrentContext(_ context.Context) (string, error) {
	return f.currentContext, nil
}
func (f fakeDiscovery) ListContexts(_ context.Context) ([]string, error) {
	return []string{f.currentContext}, nil
}
func (f fakeDiscovery) ListNamespaces(_ context.Context, _ string) ([]string, error) {
	return f.namespaces, nil
}
func (f fakeDiscovery) ListTargets(_ context.Context, _, _ string) ([]domain.Target, error) {
	return f.targets, f.targetsErr
}

type fakeStore struct {
	cfg domain.AppConfig
}

func (f *fakeStore) Load() (domain.AppConfig, error) { return f.cfg, nil }
func (f *fakeStore) Save(cfg domain.AppConfig) error { f.cfg = cfg; return nil }

func TestLoadCatalogMergesConfigAndRanksTargets(t *testing.T) {
	discovery := fakeDiscovery{
		currentContext: "dev",
		namespaces:     []string{"default", "cco"},
		targets: []domain.Target{
			{Name: "worker", Type: domain.TargetTypePod, RemotePort: 8080},
			{Name: "admin", Type: domain.TargetTypeService, RemotePort: 3000},
		},
	}
	store := &fakeStore{cfg: domain.AppConfig{
		CurrentNamespace: "cco",
		Targets: map[string]domain.TargetConfig{
			"service:admin": {Alias: "admin", PreferredLocalPort: 3001, Favorite: true},
		},
	}}

	result, err := LoadCatalog(context.Background(), Dependencies{Discovery: discovery, ConfigStore: store})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Context != "dev" || result.Namespace != "cco" {
		t.Fatalf("expected dev/cco, got %s/%s", result.Context, result.Namespace)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].Label != "admin" || result.Items[0].PreferredLocalPort != 3001 {
		t.Fatalf("expected favorite admin first with port 3001, got %+v", result.Items[0])
	}
}
