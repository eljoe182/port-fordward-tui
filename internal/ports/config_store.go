package ports

import "port-forward-tui/internal/domain"

type ConfigStore interface {
	Load() (domain.AppConfig, error)
	Save(cfg domain.AppConfig) error
}
