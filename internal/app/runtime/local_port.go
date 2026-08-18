package runtime

import (
	"fmt"

	"port-forward-tui/internal/ports"
)

func NextAvailableLocalPort(preferred int, reserved map[int]struct{}, checker ports.LocalPortChecker) (int, error) {
	if preferred < 1 || preferred > 65535 {
		return 0, fmt.Errorf("preferred local port %d must be in range 1..65535", preferred)
	}

	for port := preferred; port <= 65535; port++ {
		if _, exists := reserved[port]; exists {
			continue
		}
		if checker != nil && !checker.Available(port) {
			continue
		}
		return port, nil
	}

	return 0, fmt.Errorf("no local port available from %d to 65535", preferred)
}
