package runtime

import (
	"context"
	"testing"

	"cco-port-forward-tui/internal/domain"
)

type fakeRunner struct {
	calls []domain.ForwardRequest
}

func (f *fakeRunner) Start(_ context.Context, req domain.ForwardRequest) (string, error) {
	f.calls = append(f.calls, req)
	return req.TargetID, nil
}

func TestStartRejectsConflictingLocalPorts(t *testing.T) {
	runner := &fakeRunner{}
	svc := NewService(runner)

	selection := []domain.ForwardRequest{
		{TargetID: "svc:admin", LocalPort: 3001, RemotePort: 3000},
		{TargetID: "pod:redis", LocalPort: 3001, RemotePort: 6379},
	}

	err := svc.StartMany(context.Background(), selection)
	if err == nil {
		t.Fatalf("expected conflict error")
	}
}
