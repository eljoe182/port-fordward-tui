package catalog

import (
	"strings"

	"port-forward-tui/internal/domain"
)

func ApplyFilters(targets []domain.Target, opts LoadOptions) []domain.Target {
	filtered := make([]domain.Target, 0, len(targets))
	query := strings.TrimSpace(strings.ToLower(opts.Query))

	for _, target := range targets {
		if !matchesFilter(target, opts.Filter) {
			continue
		}
		if query != "" && !matchesQuery(target, query) {
			continue
		}
		filtered = append(filtered, target)
	}

	return filtered
}

func matchesFilter(target domain.Target, mode FilterMode) bool {
	switch mode {
	case FilterServices:
		return target.Type == domain.TargetTypeService
	case FilterPods:
		return target.Type == domain.TargetTypePod
	case FilterFavorites:
		return target.Favorite
	default:
		return true
	}
}

func matchesQuery(target domain.Target, query string) bool {
	return strings.Contains(strings.ToLower(target.Name), query) ||
		strings.Contains(strings.ToLower(target.Alias), query)
}
