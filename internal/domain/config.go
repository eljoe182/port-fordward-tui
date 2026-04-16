package domain

type AppConfig struct {
	CurrentContext   string                  `json:"currentContext,omitempty"`
	CurrentNamespace string                  `json:"currentNamespace,omitempty"`
	Targets          map[string]TargetConfig `json:"targets"`
}
