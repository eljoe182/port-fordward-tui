package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type Runner struct{}

func New() Runner { return Runner{} }

func (Runner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("%s %v: %w (stderr: %s)", name, args, err, stderr.String())
	}
	return stdout.String(), nil
}
