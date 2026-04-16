package kubectl

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"sync"

	"cco-port-forward-tui/internal/domain"
)

type CommandBuilder func(ctx context.Context, req domain.ForwardRequest) *exec.Cmd

type Runtime struct {
	build    CommandBuilder
	events   chan domain.ForwardEvent
	mu       sync.Mutex
	sessions map[string]*session
	counter  int
}

type session struct {
	targetID string
	cancel   context.CancelFunc
	done     chan struct{}
	stopping bool
}

func NewRuntime() *Runtime {
	return NewRuntimeWithBuilder(defaultCommandBuilder)
}

func NewRuntimeWithBuilder(build CommandBuilder) *Runtime {
	return &Runtime{
		build:    build,
		events:   make(chan domain.ForwardEvent, 16),
		sessions: map[string]*session{},
	}
}

func (r *Runtime) Events() <-chan domain.ForwardEvent { return r.events }

func (r *Runtime) Start(ctx context.Context, req domain.ForwardRequest) (string, error) {
	subCtx, cancel := context.WithCancel(ctx)
	cmd := r.build(subCtx, req)

	if err := cmd.Start(); err != nil {
		cancel()
		return "", fmt.Errorf("kubectl port-forward start: %w", err)
	}

	r.mu.Lock()
	r.counter++
	sessionID := fmt.Sprintf("sid-%d", r.counter)
	sess := &session{
		targetID: req.TargetID,
		cancel:   cancel,
		done:     make(chan struct{}),
	}
	r.sessions[sessionID] = sess
	r.mu.Unlock()

	go r.monitor(sessionID, sess, cmd)

	return sessionID, nil
}

func (r *Runtime) Stop(_ context.Context, sessionID string) error {
	r.mu.Lock()
	sess, ok := r.sessions[sessionID]
	if ok {
		sess.stopping = true
	}
	r.mu.Unlock()

	if !ok {
		return nil
	}
	sess.cancel()
	<-sess.done
	return nil
}

func (r *Runtime) monitor(sessionID string, sess *session, cmd *exec.Cmd) {
	err := cmd.Wait()

	r.mu.Lock()
	stopping := sess.stopping
	delete(r.sessions, sessionID)
	r.mu.Unlock()

	status := domain.ForwardStatusFailed
	errMsg := ""
	if stopping {
		status = domain.ForwardStatusStopped
	} else if err != nil {
		errMsg = err.Error()
	}

	r.events <- domain.ForwardEvent{
		SessionID: sessionID,
		TargetID:  sess.targetID,
		Status:    status,
		Err:       errMsg,
	}
	close(sess.done)
}

func defaultCommandBuilder(ctx context.Context, req domain.ForwardRequest) *exec.Cmd {
	resourceType := req.Type
	name := req.TargetID
	if parsed, ok := domain.ParseTargetKey(req.TargetID); ok {
		resourceType = parsed.Type
		name = parsed.Name
	}
	resource := string(resourceType) + "/" + name

	return exec.CommandContext(ctx, "kubectl",
		"--context", req.Context,
		"--namespace", req.Namespace,
		"port-forward",
		resource,
		strconv.Itoa(req.LocalPort)+":"+strconv.Itoa(req.RemotePort),
	)
}
