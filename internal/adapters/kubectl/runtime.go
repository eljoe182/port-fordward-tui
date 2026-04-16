package kubectl

import (
	"context"
	"fmt"

	"cco-port-forward-tui/internal/domain"
)

type Runtime struct {
	exec ExecRunner
}

func NewRuntime(exec ExecRunner) Runtime {
	return Runtime{exec: exec}
}

func (r Runtime) Start(ctx context.Context, req domain.ForwardRequest) (string, error) {
	resource := string(req.Type) + "/" + req.TargetID
	_, err := r.exec.Run(ctx, "kubectl",
		"--context", req.Context,
		"--namespace", req.Namespace,
		"port-forward",
		resource,
		fmt.Sprintf("%d:%d", req.LocalPort, req.RemotePort),
	)
	if err != nil {
		return "", err
	}
	return req.TargetID, nil
}

func (r Runtime) Stop(_ context.Context, _ string) error {
	return nil
}
