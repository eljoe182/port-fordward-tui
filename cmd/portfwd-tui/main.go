package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"port-forward-tui/internal/adapters/configfile"
	execadapter "port-forward-tui/internal/adapters/exec"
	"port-forward-tui/internal/adapters/kubectl"
	appruntime "port-forward-tui/internal/app/runtime"
	"port-forward-tui/internal/tui"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	deps, err := bootstrap()
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	program := tea.NewProgram(
		tui.NewModel(deps).WithContext(ctx),
		tea.WithAltScreen(),
		tea.WithContext(ctx),
	)
	if _, err := program.Run(); err != nil {
		log.Fatalf("tui exited: %v", err)
	}
}

func bootstrap() (tui.Dependencies, error) {
	configDir, err := resolveConfigDir()
	if err != nil {
		return tui.Dependencies{}, err
	}

	runner := execadapter.New()
	discovery := kubectl.NewDiscoveryClient(runner)
	runtime := kubectl.NewRuntime()
	runtimeApp := appruntime.NewService(runtime)
	store := configfile.NewStore(configDir)

	return tui.Dependencies{
		Discovery:   discovery,
		ConfigStore: store,
		Runtime:     runtime,
		RuntimeApp:  runtimeApp,
	}, nil
}

func resolveConfigDir() (string, error) {
	if dir := os.Getenv("PORTFWD_TUI_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}
	return filepath.Join(base, "portfwd-tui"), nil
}
