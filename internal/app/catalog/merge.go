package catalog

import "port-forward-tui/internal/domain"

func MergeTargets(discovered []domain.Target, configs map[string]domain.TargetConfig) []domain.Target {
	merged := make([]domain.Target, 0, len(discovered))
	seen := make(map[string]struct{}, len(discovered))

	for _, target := range discovered {
		target.Available = true
		key := TargetKey(target)
		if cfg, ok := configs[key]; ok {
			target = target.WithConfig(cfg)
		} else if legacyKey, ok := legacyTargetKey(target); ok {
			if cfg, ok := configs[legacyKey]; ok {
				target = target.WithConfig(cfg)
				seen[legacyKey] = struct{}{}
			}
		}
		merged = append(merged, target)
		seen[key] = struct{}{}
	}

	for key, cfg := range configs {
		if _, ok := seen[key]; ok {
			continue
		}

		base, ok := domain.ParseTargetKey(key)
		if !ok {
			continue
		}
		base.Available = false
		base = base.WithConfig(cfg)
		merged = append(merged, base)
	}

	return merged
}

func TargetKey(target domain.Target) string {
	return domain.TargetKey(target)
}

func legacyTargetKey(target domain.Target) (string, bool) {
	if target.Namespace == "" {
		return "", false
	}
	return string(target.Type) + ":" + target.Name, true
}
