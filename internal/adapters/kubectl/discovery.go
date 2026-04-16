package kubectl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cco-port-forward-tui/internal/domain"
)

type ExecRunner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

type DiscoveryClient struct {
	exec ExecRunner
}

func NewDiscoveryClient(exec ExecRunner) DiscoveryClient {
	return DiscoveryClient{exec: exec}
}

func (c DiscoveryClient) CurrentContext(ctx context.Context) (string, error) {
	out, err := c.exec.Run(ctx, "kubectl", "config", "current-context")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (c DiscoveryClient) ListContexts(ctx context.Context) ([]string, error) {
	out, err := c.exec.Run(ctx, "kubectl", "config", "get-contexts", "-o", "name")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}
	return lines, nil
}

func (c DiscoveryClient) ListNamespaces(ctx context.Context, contextName string) ([]string, error) {
	out, err := c.exec.Run(ctx, "kubectl",
		"--context", contextName,
		"get", "namespaces",
		"-o", "jsonpath={.items[*].metadata.name}",
	)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return []string{}, nil
	}
	return strings.Fields(trimmed), nil
}

type servicesPayload struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Spec struct {
			Ports []struct {
				Port int `json:"port"`
			} `json:"ports"`
		} `json:"spec"`
	} `json:"items"`
}

type podsPayload struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Spec struct {
			Containers []struct {
				Ports []struct {
					ContainerPort int `json:"containerPort"`
				} `json:"ports"`
			} `json:"containers"`
		} `json:"spec"`
	} `json:"items"`
}

func (c DiscoveryClient) ListTargets(ctx context.Context, contextName, namespace string) ([]domain.Target, error) {
	services, err := c.listServices(ctx, contextName, namespace)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	pods, err := c.listPods(ctx, contextName, namespace)
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}
	return append(services, pods...), nil
}

func (c DiscoveryClient) listServices(ctx context.Context, contextName, namespace string) ([]domain.Target, error) {
	out, err := c.exec.Run(ctx, "kubectl",
		"--context", contextName,
		"--namespace", namespace,
		"get", "services",
		"-o", "json",
	)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}
	var payload servicesPayload
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		return nil, fmt.Errorf("decode services json: %w", err)
	}
	targets := make([]domain.Target, 0, len(payload.Items))
	for _, item := range payload.Items {
		if len(item.Spec.Ports) == 0 {
			continue
		}
		targets = append(targets, domain.Target{
			Name:       item.Metadata.Name,
			Type:       domain.TargetTypeService,
			RemotePort: item.Spec.Ports[0].Port,
		})
	}
	return targets, nil
}

func (c DiscoveryClient) listPods(ctx context.Context, contextName, namespace string) ([]domain.Target, error) {
	out, err := c.exec.Run(ctx, "kubectl",
		"--context", contextName,
		"--namespace", namespace,
		"get", "pods",
		"-o", "json",
	)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}
	var payload podsPayload
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		return nil, fmt.Errorf("decode pods json: %w", err)
	}
	targets := make([]domain.Target, 0, len(payload.Items))
	for _, item := range payload.Items {
		port := firstContainerPort(item.Spec.Containers)
		if port == 0 {
			continue
		}
		targets = append(targets, domain.Target{
			Name:       item.Metadata.Name,
			Type:       domain.TargetTypePod,
			RemotePort: port,
		})
	}
	return targets, nil
}

func firstContainerPort(containers []struct {
	Ports []struct {
		ContainerPort int `json:"containerPort"`
	} `json:"ports"`
}) int {
	for _, c := range containers {
		for _, p := range c.Ports {
			if p.ContainerPort != 0 {
				return p.ContainerPort
			}
		}
	}
	return 0
}
