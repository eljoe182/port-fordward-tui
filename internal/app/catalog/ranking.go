package catalog

import (
	"sort"
	"strings"
	"time"

	"cco-port-forward-tui/internal/domain"
)

func RankSmart(targets []domain.Target, now time.Time, query string) []domain.Target {
	type scored struct {
		target domain.Target
		score  int
	}

	items := make([]scored, 0, len(targets))
	for _, target := range targets {
		score := 0
		if target.Favorite {
			score += 100
		}
		if target.PreferredLocalPort != 0 {
			score += 25
		}
		if !target.LastUsedAt.IsZero() && now.Sub(target.LastUsedAt) < 24*time.Hour {
			score += 50
		}
		if query != "" && strings.Contains(strings.ToLower(target.Name), strings.ToLower(query)) {
			score += 75
		}
		items = append(items, scored{target: target, score: score})
	}

	sort.SliceStable(items, func(i, j int) bool { return items[i].score > items[j].score })

	result := make([]domain.Target, 0, len(items))
	for _, item := range items {
		result = append(result, item.target)
	}
	return result
}
