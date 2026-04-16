package domain

import "testing"

func TestTargetMergeConfiguredFieldsOverridesDiscoveredDefaults(t *testing.T) {
	discovered := Target{Name: "cco-admin-api", Type: TargetTypeService, RemotePort: 3000}
	configured := TargetConfig{Alias: "admin", PreferredLocalPort: 3001, Favorite: true}

	merged := discovered.WithConfig(configured)

	if merged.Alias != "admin" || merged.PreferredLocalPort != 3001 || !merged.Favorite {
		t.Fatalf("expected configured values to override discovered defaults: %+v", merged)
	}
}
