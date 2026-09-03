package azure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

type scriptedCall struct {
	name   string
	args   []string
	output string
	err    error
}

type scriptedExec struct {
	t     *testing.T
	calls []scriptedCall
	idx   int
	seen  [][]string
}

func (s *scriptedExec) Run(_ context.Context, name string, args ...string) (string, error) {
	s.t.Helper()
	full := append([]string{name}, args...)
	s.seen = append(s.seen, full)

	if s.idx >= len(s.calls) {
		s.t.Fatalf("unexpected extra call: %v", full)
	}
	want := s.calls[s.idx]
	s.idx++

	if name != want.name {
		s.t.Fatalf("call %d: got command %q, want %q", s.idx, name, want.name)
	}
	if len(args) != len(want.args) {
		s.t.Fatalf("call %d: got args %#v, want %#v", s.idx, args, want.args)
	}
	for i := range args {
		if args[i] != want.args[i] {
			s.t.Fatalf("call %d: got args %#v, want %#v", s.idx, args, want.args)
		}
	}
	return want.output, want.err
}

func TestSyncSuccessAcrossSubscriptions(t *testing.T) {
	exec := &scriptedExec{
		t: t,
		calls: []scriptedCall{
			{name: "az", args: []string{"account", "list", "--query", "[].id", "-o", "tsv"}, output: "sub-a\nsub-b\n"},
			{name: "az", args: []string{"aks", "list", "--subscription", "sub-a", "--query", "[].[name, resourceGroup]", "-o", "tsv"}, output: "cluster-a\trg-a\n"},
			{name: "az", args: []string{"aks", "get-credentials", "--resource-group", "rg-a", "--name", "cluster-a", "--subscription", "sub-a", "--overwrite-existing"}},
			{name: "az", args: []string{"aks", "list", "--subscription", "sub-b", "--query", "[].[name, resourceGroup]", "-o", "tsv"}, output: "cluster-b\trg-b\n"},
			{name: "az", args: []string{"aks", "get-credentials", "--resource-group", "rg-b", "--name", "cluster-b", "--subscription", "sub-b", "--overwrite-existing"}},
		},
	}

	result, err := NewAKSSyncer(exec).Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Provider != "Azure AKS" {
		t.Fatalf("unexpected provider: %q", result.Provider)
	}
	if result.Synced != 2 {
		t.Fatalf("expected Synced=2, got %d", result.Synced)
	}
	if len(result.Failed) != 0 {
		t.Fatalf("expected no failures, got %#v", result.Failed)
	}
	if exec.idx != len(exec.calls) {
		t.Fatalf("expected %d calls, made %d", len(exec.calls), exec.idx)
	}
}

func TestSyncPartialClusterFailureContinues(t *testing.T) {
	exec := &scriptedExec{
		t: t,
		calls: []scriptedCall{
			{name: "az", args: []string{"account", "list", "--query", "[].id", "-o", "tsv"}, output: "sub-a\n"},
			{name: "az", args: []string{"aks", "list", "--subscription", "sub-a", "--query", "[].[name, resourceGroup]", "-o", "tsv"}, output: "bad\trg-bad\ngood\trg-good\n"},
			{
				name: "az",
				args: []string{"aks", "get-credentials", "--resource-group", "rg-bad", "--name", "bad", "--subscription", "sub-a", "--overwrite-existing"},
				err:  errors.New("permission denied"),
			},
			{name: "az", args: []string{"aks", "get-credentials", "--resource-group", "rg-good", "--name", "good", "--subscription", "sub-a", "--overwrite-existing"}},
		},
	}

	result, err := NewAKSSyncer(exec).Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Synced != 1 {
		t.Fatalf("expected Synced=1, got %d", result.Synced)
	}
	if len(result.Failed) != 1 || result.Failed[0].Name != "rg-bad/bad" {
		t.Fatalf("unexpected failures: %#v", result.Failed)
	}
	if !strings.Contains(result.Failed[0].Detail, "permission denied") {
		t.Fatalf("expected failure detail to include permission denied, got %q", result.Failed[0].Detail)
	}
}

func TestSyncAccountListFailureIsHardError(t *testing.T) {
	exec := &scriptedExec{
		t: t,
		calls: []scriptedCall{
			{
				name: "az",
				args: []string{"account", "list", "--query", "[].id", "-o", "tsv"},
				err:  fmt.Errorf("az: not logged in"),
			},
		},
	}

	_, err := NewAKSSyncer(exec).Sync(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "list Azure subscriptions") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncEmptySubscriptions(t *testing.T) {
	exec := &scriptedExec{
		t: t,
		calls: []scriptedCall{
			{name: "az", args: []string{"account", "list", "--query", "[].id", "-o", "tsv"}, output: "\n"},
		},
	}

	result, err := NewAKSSyncer(exec).Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Synced != 0 || len(result.Failed) != 0 {
		t.Fatalf("expected empty result, got %#v", result)
	}
}

func TestProviderName(t *testing.T) {
	if got := NewAKSSyncer(nil).Provider(); got != "Azure AKS" {
		t.Fatalf("unexpected provider: %q", got)
	}
}
