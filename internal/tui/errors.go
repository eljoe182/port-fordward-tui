package tui

import "strings"

func actionableError(err string) string {
	trimmed := strings.TrimSpace(err)
	if trimmed == "" {
		return ""
	}
	lower := strings.ToLower(trimmed)
	switch {
	case strings.Contains(lower, "address already in use") || strings.Contains(lower, "port already selected") || strings.Contains(lower, "port in use") || strings.Contains(lower, "bind"):
		return "local port unavailable — edit the local port and retry"
	case strings.Contains(lower, "not found"):
		return "target not found — refresh catalog or verify context/namespace"
	case strings.Contains(lower, "connection refused"), strings.Contains(lower, "unable to listen"):
		return "port-forward failed to bind — try a different local port"
	case strings.Contains(lower, "forbidden"), strings.Contains(lower, "unauthorized"):
		return "access denied — verify kubectl credentials and cluster permissions"
	default:
		return trimmed
	}
}
