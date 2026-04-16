package kubectl

import (
	"context"
	"strings"
	"testing"

	"cco-port-forward-tui/internal/domain"
)

type scriptedExec struct {
	responses map[string]string
}

func (s scriptedExec) Run(_ context.Context, name string, args ...string) (string, error) {
	key := name + " " + strings.Join(args, " ")
	if out, ok := s.responses[key]; ok {
		return out, nil
	}
	return "", nil
}

const servicesJSON = `{
  "items": [
    {
      "metadata": {"name": "cco-admin-api"},
      "spec": {"ports": [{"port": 3000, "targetPort": 3000}]}
    },
    {
      "metadata": {"name": "cco-db"},
      "spec": {"ports": [{"port": 5432, "targetPort": 5432}]}
    }
  ]
}`

const podsJSON = `{
  "items": [
    {
      "metadata": {"name": "redis-0"},
      "spec": {
        "containers": [
          {"name": "redis", "ports": [{"containerPort": 6379}]}
        ]
      }
    },
    {
      "metadata": {"name": "no-ports"},
      "spec": {"containers": [{"name": "sidecar"}]}
    }
  ]
}`

func TestListTargetsMergesServicesAndPodsWithPorts(t *testing.T) {
	exec := scriptedExec{
		responses: map[string]string{
			"kubectl --context dev --namespace cco get services -o json": servicesJSON,
			"kubectl --context dev --namespace cco get pods -o json":     podsJSON,
		},
	}
	client := NewDiscoveryClient(exec)

	targets, err := client.ListTargets(context.Background(), "dev", "cco")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	byName := map[string]domain.Target{}
	for _, target := range targets {
		byName[target.Name] = target
	}

	admin, ok := byName["cco-admin-api"]
	if !ok || admin.Type != domain.TargetTypeService || admin.RemotePort != 3000 || admin.Namespace != "cco" || !admin.Available {
		t.Fatalf("expected service cco-admin-api with port 3000, got %+v", admin)
	}
	redis, ok := byName["redis-0"]
	if !ok || redis.Type != domain.TargetTypePod || redis.RemotePort != 6379 || redis.Namespace != "cco" || !redis.Available {
		t.Fatalf("expected pod redis-0 with port 6379, got %+v", redis)
	}
	if _, present := byName["no-ports"]; present {
		t.Fatalf("pods without declared ports must be skipped")
	}
}
