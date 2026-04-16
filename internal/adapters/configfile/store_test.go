package configfile

import (
	"testing"

	"port-forward-tui/internal/domain"
)

func TestStoreRoundTripPersistsTargetConfig(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	config := domain.AppConfig{
		Targets: map[string]domain.TargetConfig{
			"service:cco:admin": {Alias: "admin", PreferredLocalPort: 3001, Favorite: true},
		},
	}

	if err := store.Save(config); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Targets["service:cco:admin"].Alias != "admin" {
		t.Fatalf("expected alias persisted")
	}
}
