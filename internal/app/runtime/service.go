package runtime

import (
	"context"
	"fmt"

	"port-forward-tui/internal/domain"
	"port-forward-tui/internal/ports"
)

type StartResult struct {
	Request   domain.ForwardRequest
	SessionID string
	Err       error
}

type Service struct {
	runner ports.ForwardRunner
}

func NewService(runner ports.ForwardRunner) Service {
	return Service{runner: runner}
}

func (s Service) Available() bool { return s.runner != nil }

func (s Service) StartMany(ctx context.Context, requests []domain.ForwardRequest, active []domain.ForwardSession) ([]StartResult, error) {
	if s.runner == nil {
		return nil, fmt.Errorf("runtime service unavailable")
	}
	if err := ValidateRequests(requests, active); err != nil {
		return nil, err
	}
	results := make([]StartResult, 0, len(requests))
	for _, req := range requests {
		result := StartResult{Request: req}
		sessionID, err := s.runner.Start(ctx, req)
		if err != nil {
			result.Err = err
		} else {
			result.SessionID = sessionID
		}
		results = append(results, result)
	}
	return results, nil
}

func (s Service) StartOne(ctx context.Context, request domain.ForwardRequest, active []domain.ForwardSession) (StartResult, error) {
	if s.runner == nil {
		return StartResult{}, fmt.Errorf("runtime service unavailable")
	}
	if err := ValidateRequests([]domain.ForwardRequest{request}, active); err != nil {
		return StartResult{}, err
	}
	sessionID, err := s.runner.Start(ctx, request)
	if err != nil {
		return StartResult{Request: request, Err: err}, nil
	}
	return StartResult{Request: request, SessionID: sessionID}, nil
}

func ValidateRequests(requests []domain.ForwardRequest, active []domain.ForwardSession) error {
	seen := map[int]string{}
	for _, session := range active {
		if session.Status == domain.ForwardStatusStopped || session.Status == domain.ForwardStatusFailed {
			continue
		}
		seen[session.LocalPort] = session.TargetID
	}

	for _, req := range requests {
		if req.LocalPort < 1 || req.LocalPort > 65535 {
			return fmt.Errorf("local port %d must be in range 1..65535", req.LocalPort)
		}
		if existing, exists := seen[req.LocalPort]; exists && existing != req.TargetID {
			return fmt.Errorf("local port %d already in use by %s", req.LocalPort, existing)
		}
		seen[req.LocalPort] = req.TargetID
	}
	return nil
}
