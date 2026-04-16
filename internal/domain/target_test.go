package domain

import (
	"testing"
	"time"
)

func TestTargetMergeConfiguredFieldsOverridesDiscoveredDefaults(t *testing.T) {
	discovered := Target{Name: "cco-admin-api", Type: TargetTypeService, RemotePort: 3000}
	configured := TargetConfig{Alias: "admin", PreferredLocalPort: 3001, Favorite: true}

	merged := discovered.WithConfig(configured)

	if merged.Alias != "admin" || merged.PreferredLocalPort != 3001 || !merged.Favorite {
		t.Fatalf("expected configured values to override discovered defaults: %+v", merged)
	}
}

func TestTargetKeyIncludesNamespaceWhenPresent(t *testing.T) {
	target := Target{Type: TargetTypeService, Namespace: "cco", Name: "admin"}

	if got := TargetKey(target); got != "service:cco:admin" {
		t.Fatalf("expected namespaced key, got %q", got)
	}
}

func TestParseTargetKeySupportsLegacyAndNamespacedKeys(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want Target
	}{
		{
			name: "legacy key without namespace",
			key:  "service:admin",
			want: Target{Type: TargetTypeService, Name: "admin", Available: true},
		},
		{
			name: "namespaced key",
			key:  "pod:cco:redis",
			want: Target{Type: TargetTypePod, Namespace: "cco", Name: "redis", Available: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseTargetKey(tt.key)
			if !ok {
				t.Fatalf("expected key %q to parse", tt.key)
			}
			if got != tt.want {
				t.Fatalf("unexpected parse result: got %+v want %+v", got, tt.want)
			}
		})
	}
}

func TestTargetToConfigCarriesPersistentMetadata(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	target := Target{
		Type:               TargetTypeService,
		Namespace:          "cco",
		Name:               "admin",
		Alias:              "admin",
		PreferredLocalPort: 3001,
		Favorite:           true,
		RemotePort:         3000,
		LastUsedAt:         now,
	}

	got := target.ToConfig()
	if got.Type != TargetTypeService || got.Namespace != "cco" || got.Name != "admin" || got.RemotePort != 3000 || !got.LastUsedAt.Equal(now) {
		t.Fatalf("unexpected config conversion: %+v", got)
	}
}
