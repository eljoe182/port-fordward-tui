package domain

import "time"

type TargetType string

const (
	TargetTypeService TargetType = "service"
	TargetTypePod     TargetType = "pod"
)

type Target struct {
	Name               string
	Type               TargetType
	Alias              string
	RemotePort         int
	PreferredLocalPort int
	Favorite           bool
	LastUsedAt         time.Time
}

type TargetConfig struct {
	Alias              string `json:"alias"`
	PreferredLocalPort int    `json:"preferredLocalPort"`
	Favorite           bool   `json:"favorite"`
}

func (t Target) WithConfig(cfg TargetConfig) Target {
	if cfg.Alias != "" {
		t.Alias = cfg.Alias
	}
	if cfg.PreferredLocalPort != 0 {
		t.PreferredLocalPort = cfg.PreferredLocalPort
	}
	t.Favorite = cfg.Favorite
	return t
}
