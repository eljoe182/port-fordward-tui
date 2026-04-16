package catalog

import (
	"testing"
	"time"

	"port-forward-tui/internal/domain"
)

func TestSmartRankingPrioritizesFavoriteRecentConfiguredTargets(t *testing.T) {
	now := time.Now()
	targets := []domain.Target{
		{Name: "worker", Type: domain.TargetTypePod, Available: true},
		{Name: "admin", Type: domain.TargetTypeService, Favorite: true, PreferredLocalPort: 3001, Available: true},
		{Name: "redis", Type: domain.TargetTypePod, LastUsedAt: now.Add(-1 * time.Hour), Available: true},
	}

	ranked := RankSmart(targets, now, "")

	if ranked[0].Name != "admin" {
		t.Fatalf("expected favorite configured target first, got %s", ranked[0].Name)
	}
}

func TestMergeTargetsAppendsConfiguredTargetMissingFromDiscovery(t *testing.T) {
	discovered := []domain.Target{{Type: domain.TargetTypeService, Namespace: "cco", Name: "admin", RemotePort: 3000, Available: true}}
	configs := map[string]domain.TargetConfig{
		"service:cco:admin": {Alias: "admin", PreferredLocalPort: 3001, Favorite: true, Type: domain.TargetTypeService, Namespace: "cco", Name: "admin", RemotePort: 3000},
		"pod:cco:redis":     {Alias: "redis", PreferredLocalPort: 54724, Favorite: true, Type: domain.TargetTypePod, Namespace: "cco", Name: "redis", RemotePort: 6379},
	}

	merged := MergeTargets(discovered, configs)
	if len(merged) != 2 {
		t.Fatalf("expected 2 merged targets, got %d", len(merged))
	}

	var stale domain.Target
	for _, target := range merged {
		if target.Name == "redis" {
			stale = target
		}
	}

	if stale.Name != "redis" || stale.Available {
		t.Fatalf("expected configured-only redis target marked unavailable, got %+v", stale)
	}
	if stale.PreferredLocalPort != 54724 || stale.RemotePort != 6379 {
		t.Fatalf("expected configured metadata preserved, got %+v", stale)
	}
}

func TestMergeTargetsSupportsLegacyConfigKeysForDiscoveredTargets(t *testing.T) {
	discovered := []domain.Target{{Type: domain.TargetTypeService, Namespace: "cco", Name: "admin", RemotePort: 3000, Available: true}}
	configs := map[string]domain.TargetConfig{
		"service:admin": {Alias: "legacy-admin", PreferredLocalPort: 3009, Favorite: true},
	}

	merged := MergeTargets(discovered, configs)
	if len(merged) != 1 {
		t.Fatalf("expected 1 merged target, got %d", len(merged))
	}
	if merged[0].Alias != "legacy-admin" || merged[0].PreferredLocalPort != 3009 {
		t.Fatalf("expected legacy config applied to discovered target, got %+v", merged[0])
	}
}

func TestApplyFiltersUsesQueryAndFavoritesMode(t *testing.T) {
	targets := []domain.Target{
		{Name: "admin", Alias: "dashboard", Favorite: true, Type: domain.TargetTypeService},
		{Name: "redis", Type: domain.TargetTypePod},
	}

	filtered := ApplyFilters(targets, LoadOptions{Query: "dash", Filter: FilterFavorites})
	if len(filtered) != 1 || filtered[0].Name != "admin" {
		t.Fatalf("expected only favorite query match, got %+v", filtered)
	}
}

func TestSortTargetsByName(t *testing.T) {
	targets := []domain.Target{
		{Name: "redis", Type: domain.TargetTypePod},
		{Name: "admin", Type: domain.TargetTypeService},
	}

	sorted := SortTargets(targets, time.Now(), LoadOptions{Sort: SortName})
	if sorted[0].Name != "admin" {
		t.Fatalf("expected alphabetical sort, got %+v", sorted)
	}
}
