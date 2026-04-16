package kubectl

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"sync"

	"cco-port-forward-tui/internal/domain"
)

type Runtime struct {
	mu       sync.Mutex
	sessions map[string]*session
	counter  int
}

type session struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
	done   chan error
}

func NewRuntime() *Runtime {
	return &Runtime{sessions: map[string]*session{}}
}

func (r *Runtime) Start(ctx context.Context, req domain.ForwardRequest) (string, error) {
	subCtx, cancel := context.WithCancel(ctx)

	resource := string(req.Type) + "/" + req.TargetID
	if colonIdx := indexOf(req.TargetID, ':'); colonIdx > 0 {
		resource = string(req.Type) + "/" + req.TargetID[colonIdx+1:]
	}

	cmd := exec.CommandContext(subCtx, "kubectl",
		"--context", req.Context,
		"--namespace", req.Namespace,
		"port-forward",
		resource,
		strconv.Itoa(req.LocalPort)+":"+strconv.Itoa(req.RemotePort),
	)

	if err := cmd.Start(); err != nil {
		cancel()
		return "", fmt.Errorf("kubectl port-forward start: %w", err)
	}

	r.mu.Lock()
	r.counter++
	sessionID := fmt.Sprintf("sid-%d", r.counter)
	done := make(chan error, 1)
	r.sessions[sessionID] = &session{cmd: cmd, cancel: cancel, done: done}
	r.mu.Unlock()

	go func() {
		done <- cmd.Wait()
		close(done)
	}()

	return sessionID, nil
}

func (r *Runtime) Stop(_ context.Context, sessionID string) error {
	r.mu.Lock()
	sess, ok := r.sessions[sessionID]
	if ok {
		delete(r.sessions, sessionID)
	}
	r.mu.Unlock()

	if !ok {
		return nil
	}
	sess.cancel()
	<-sess.done
	return nil
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
