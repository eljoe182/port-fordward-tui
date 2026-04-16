package kubectl

import (
	"context"
	"testing"
)

type fakeExec struct {
	output string
}

func (f fakeExec) Run(_ context.Context, _ string, _ ...string) (string, error) {
	return f.output, nil
}

func TestListContextsParsesKubectlOutput(t *testing.T) {
	exec := fakeExec{output: "dev\nprod\n"}
	client := NewDiscoveryClient(exec)

	contexts, err := client.ListContexts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(contexts) != 2 || contexts[0] != "dev" || contexts[1] != "prod" {
		t.Fatalf("unexpected contexts: %#v", contexts)
	}
}
