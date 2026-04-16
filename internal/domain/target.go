package domain

import (
	"strings"
	"time"
)

type TargetType string

const (
	TargetTypeService TargetType = "service"
	TargetTypePod     TargetType = "pod"
)

type Target struct {
	Namespace          string
	Name               string
	Type               TargetType
	Alias              string
	RemotePort         int
	PreferredLocalPort int
	Favorite           bool
	LastUsedAt         time.Time
	Available          bool
}

type TargetConfig struct {
	Type               TargetType `json:"type,omitempty"`
	Namespace          string     `json:"namespace,omitempty"`
	Name               string     `json:"name,omitempty"`
	Alias              string     `json:"alias"`
	PreferredLocalPort int        `json:"preferredLocalPort"`
	Favorite           bool       `json:"favorite"`
	RemotePort         int        `json:"remotePort,omitempty"`
	LastUsedAt         time.Time  `json:"lastUsedAt,omitempty"`
}

func (t Target) WithConfig(cfg TargetConfig) Target {
	if cfg.Type != "" {
		t.Type = cfg.Type
	}
	if cfg.Namespace != "" {
		t.Namespace = cfg.Namespace
	}
	if cfg.Name != "" {
		t.Name = cfg.Name
	}
	if cfg.Alias != "" {
		t.Alias = cfg.Alias
	}
	if cfg.PreferredLocalPort != 0 {
		t.PreferredLocalPort = cfg.PreferredLocalPort
	}
	if cfg.RemotePort != 0 {
		t.RemotePort = cfg.RemotePort
	}
	if !cfg.LastUsedAt.IsZero() {
		t.LastUsedAt = cfg.LastUsedAt
	}
	t.Favorite = cfg.Favorite
	return t
}

func (t Target) ToConfig() TargetConfig {
	return TargetConfig{
		Type:               t.Type,
		Namespace:          t.Namespace,
		Name:               t.Name,
		Alias:              t.Alias,
		PreferredLocalPort: t.PreferredLocalPort,
		Favorite:           t.Favorite,
		RemotePort:         t.RemotePort,
		LastUsedAt:         t.LastUsedAt,
	}
}

func TargetKey(target Target) string {
	parts := []string{string(target.Type)}
	if target.Namespace != "" {
		parts = append(parts, target.Namespace)
	}
	parts = append(parts, target.Name)
	return strings.Join(parts, ":")
}

func ParseTargetKey(key string) (Target, bool) {
	parts := strings.Split(key, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return Target{}, false
	}
	target := Target{Type: TargetType(parts[0]), Available: true}
	if len(parts) == 2 {
		target.Name = parts[1]
		return target, target.Type != "" && target.Name != ""
	}
	target.Namespace = parts[1]
	target.Name = parts[2]
	return target, target.Type != "" && target.Name != ""
}
