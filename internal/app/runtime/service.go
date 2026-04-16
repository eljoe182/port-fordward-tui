package runtime

import (
	"context"
	"fmt"

	"cco-port-forward-tui/internal/domain"
)

type Runner interface {
	Start(ctx context.Context, req domain.ForwardRequest) (string, error)
}

type Service struct {
	runner Runner
}

func NewService(runner Runner) Service {
	return Service{runner: runner}
}

func (s Service) StartMany(ctx context.Context, requests []domain.ForwardRequest) error {
	seen := map[int]struct{}{}
	for _, req := range requests {
		if _, exists := seen[req.LocalPort]; exists {
			return fmt.Errorf("local port %d already selected", req.LocalPort)
		}
		seen[req.LocalPort] = struct{}{}
	}
	for _, req := range requests {
		if _, err := s.runner.Start(ctx, req); err != nil {
			return err
		}
	}
	return nil
}
