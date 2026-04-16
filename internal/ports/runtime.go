package ports

import (
	"context"

	"cco-port-forward-tui/internal/domain"
)

type ForwardRunner interface {
	Start(ctx context.Context, req domain.ForwardRequest) (sessionID string, err error)
	Stop(ctx context.Context, sessionID string) error
}
