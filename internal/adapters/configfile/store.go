package configfile

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"port-forward-tui/internal/domain"
)

type Store struct {
	path string
}

func NewStore(baseDir string) Store {
	return Store{path: filepath.Join(baseDir, "config.json")}
}

func (s Store) Save(cfg domain.AppConfig) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s Store) Load() (domain.AppConfig, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return domain.AppConfig{Targets: map[string]domain.TargetConfig{}}, nil
	}
	if err != nil {
		return domain.AppConfig{}, err
	}
	var cfg domain.AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return domain.AppConfig{}, err
	}
	if cfg.Targets == nil {
		cfg.Targets = map[string]domain.TargetConfig{}
	}
	return cfg, nil
}
