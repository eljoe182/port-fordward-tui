package catalog

import "cco-port-forward-tui/internal/domain"

func MergeTargets(discovered []domain.Target, configs map[string]domain.TargetConfig) []domain.Target {
	if len(configs) == 0 {
		return discovered
	}

	merged := make([]domain.Target, 0, len(discovered))
	seen := make(map[string]struct{}, len(discovered))

	for _, target := range discovered {
		key := TargetKey(target)
		if cfg, ok := configs[key]; ok {
			target = target.WithConfig(cfg)
		}
		merged = append(merged, target)
		seen[key] = struct{}{}
	}

	return merged
}

func TargetKey(target domain.Target) string {
	return string(target.Type) + ":" + target.Name
}
