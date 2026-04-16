package catalog

import (
	"testing"
	"time"

	"cco-port-forward-tui/internal/domain"
)

func TestSmartRankingPrioritizesFavoriteRecentConfiguredTargets(t *testing.T) {
	now := time.Now()
	targets := []domain.Target{
		{Name: "worker", Type: domain.TargetTypePod},
		{Name: "admin", Type: domain.TargetTypeService, Favorite: true, PreferredLocalPort: 3001},
		{Name: "redis", Type: domain.TargetTypePod, LastUsedAt: now.Add(-1 * time.Hour)},
	}

	ranked := RankSmart(targets, now, "")

	if ranked[0].Name != "admin" {
		t.Fatalf("expected favorite configured target first, got %s", ranked[0].Name)
	}
}
