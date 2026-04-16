package runtime

import (
	"context"
	"errors"
	"testing"

	"cco-port-forward-tui/internal/domain"
)

type fakeRunner struct {
	calls []domain.ForwardRequest
	err   error
}

func (f *fakeRunner) Start(_ context.Context, req domain.ForwardRequest) (string, error) {
	f.calls = append(f.calls, req)
	if f.err != nil {
		return "", f.err
	}
	return req.TargetID + ":session", nil
}

func (f *fakeRunner) Stop(_ context.Context, _ string) error { return nil }
func (f *fakeRunner) Events() <-chan domain.ForwardEvent     { return nil }

func TestValidateRequestsRejectsConflictingLocalPorts(t *testing.T) {
	selection := []domain.ForwardRequest{
		{TargetID: "svc:admin", LocalPort: 3001, RemotePort: 3000},
		{TargetID: "pod:redis", LocalPort: 3001, RemotePort: 6379},
	}

	err := ValidateRequests(selection, nil)
	if err == nil {
		t.Fatalf("expected conflict error")
	}
}

func TestValidateRequestsRejectsConflictWithActiveSession(t *testing.T) {
	selection := []domain.ForwardRequest{{TargetID: "svc:admin", LocalPort: 3001, RemotePort: 3000}}
	active := []domain.ForwardSession{{TargetID: "svc:other", LocalPort: 3001, Status: domain.ForwardStatusRunning}}

	err := ValidateRequests(selection, active)
	if err == nil {
		t.Fatalf("expected active session conflict")
	}
}

func TestStartManyReturnsResultPerRequest(t *testing.T) {
	runner := &fakeRunner{}
	svc := NewService(runner)
	selection := []domain.ForwardRequest{{TargetID: "svc:admin", LocalPort: 3001, RemotePort: 3000}}

	results, err := svc.StartMany(context.Background(), selection, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].SessionID == "" {
		t.Fatalf("expected successful start result, got %+v", results)
	}
}

func TestStartOneReturnsPerRequestFailureWithoutBubbling(t *testing.T) {
	runner := &fakeRunner{err: errors.New("boom")}
	svc := NewService(runner)

	result, err := svc.StartOne(context.Background(), domain.ForwardRequest{TargetID: "svc:admin", LocalPort: 3001}, nil)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if result.Err == nil || result.SessionID != "" {
		t.Fatalf("expected start failure captured in result, got %+v", result)
	}
}
