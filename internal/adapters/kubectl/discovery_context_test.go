package kubectl

import (
	"context"
	"testing"
)

func TestCurrentContextReturnsTrimmedOutput(t *testing.T) {
	exec := fakeExec{output: "dev\n"}
	client := NewDiscoveryClient(exec)

	current, err := client.CurrentContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if current != "dev" {
		t.Fatalf("expected 'dev', got %q", current)
	}
}

func TestListNamespacesParsesJsonpathOutput(t *testing.T) {
	exec := fakeExec{output: "default cco staging"}
	client := NewDiscoveryClient(exec)

	namespaces, err := client.ListNamespaces(context.Background(), "dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(namespaces) != 3 || namespaces[1] != "cco" {
		t.Fatalf("unexpected namespaces: %#v", namespaces)
	}
}
