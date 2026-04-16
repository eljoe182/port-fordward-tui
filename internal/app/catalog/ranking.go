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
		if target.Available {
			score += 20
		} else {
			score -= 40
		}
		if target.Favorite {
			score += 100
		}
		if target.PreferredLocalPort != 0 {
			score += 25
		}
		if !target.LastUsedAt.IsZero() && now.Sub(target.LastUsedAt) < 24*time.Hour {
			score += 50
		}
		if query != "" {
			q := strings.ToLower(query)
			if strings.Contains(strings.ToLower(target.Name), q) {
				score += 75
			}
			if target.Alias != "" && strings.Contains(strings.ToLower(target.Alias), q) {
				score += 50
			}
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

func SortTargets(targets []domain.Target, now time.Time, opts LoadOptions) []domain.Target {
	opts = opts.WithDefaults()
	items := append([]domain.Target(nil), targets...)

	switch opts.Sort {
	case SortName:
		sort.SliceStable(items, func(i, j int) bool {
			return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		})
	case SortRecent:
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].LastUsedAt.After(items[j].LastUsedAt)
		})
	case SortFavorites:
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].Favorite == items[j].Favorite {
				return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
			}
			return items[i].Favorite
		})
	case SortType:
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].Type == items[j].Type {
				return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
			}
			return items[i].Type < items[j].Type
		})
	default:
		return RankSmart(items, now, opts.Query)
	}

	return items
}
