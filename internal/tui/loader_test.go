package tui

import (
	"context"
	"testing"

	"cco-port-forward-tui/internal/app/catalog"
	"cco-port-forward-tui/internal/domain"
)

type fakeDiscovery struct {
	currentContext string
	contexts       []string
	namespaces     []string
	targets        []domain.Target
	targetsErr     error
}

func (f fakeDiscovery) CurrentContext(_ context.Context) (string, error) {
	return f.currentContext, nil
}
func (f fakeDiscovery) ListContexts(_ context.Context) ([]string, error) {
	if len(f.contexts) > 0 {
		return f.contexts, nil
	}
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
			{Name: "worker", Namespace: "cco", Type: domain.TargetTypePod, RemotePort: 8080, Available: true},
			{Name: "admin", Namespace: "cco", Type: domain.TargetTypeService, RemotePort: 3000, Available: true},
		},
	}
	store := &fakeStore{cfg: domain.AppConfig{
		CurrentNamespace: "cco",
		Targets: map[string]domain.TargetConfig{
			"service:cco:admin": {Alias: "admin", PreferredLocalPort: 3001, Favorite: true, Type: domain.TargetTypeService, Namespace: "cco", Name: "admin", RemotePort: 3000},
			"pod:cco:redis":     {Alias: "redis", PreferredLocalPort: 54724, Favorite: true, Type: domain.TargetTypePod, Namespace: "cco", Name: "redis", RemotePort: 6379},
		},
	}}

	result, err := LoadCatalog(context.Background(), Dependencies{Discovery: discovery, ConfigStore: store}, catalog.LoadOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Context != "dev" || result.Namespace != "cco" {
		t.Fatalf("expected dev/cco, got %s/%s", result.Context, result.Namespace)
	}
	if len(result.Contexts) != 1 || result.Contexts[0] != "dev" {
		t.Fatalf("expected contexts to be loaded, got %#v", result.Contexts)
	}
	if len(result.Namespaces) != 2 || result.Namespaces[1] != "cco" {
		t.Fatalf("expected namespaces to be loaded, got %#v", result.Namespaces)
	}
	if len(result.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result.Items))
	}
	if result.Items[0].Label != "admin" || result.Items[0].PreferredLocalPort != 3001 {
		t.Fatalf("expected favorite admin first with port 3001, got %+v", result.Items[0])
	}
	foundRedis := false
	for _, item := range result.Items {
		if item.Name == "redis" && !item.Available {
			foundRedis = true
		}
	}
	if !foundRedis {
		t.Fatalf("expected configured-only redis target included, got %#v", result.Items)
	}
}
