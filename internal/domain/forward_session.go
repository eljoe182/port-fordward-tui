package domain

import "time"

type ForwardStatus string

const (
	ForwardStatusStarting ForwardStatus = "starting"
	ForwardStatusRunning  ForwardStatus = "running"
	ForwardStatusStopped  ForwardStatus = "stopped"
	ForwardStatusFailed   ForwardStatus = "failed"
)

type ForwardRequest struct {
	TargetID   string
	Label      string
	LocalPort  int
	RemotePort int
	Context    string
	Namespace  string
	Type       TargetType
}

type ForwardSession struct {
	TargetID   string
	Label      string
	LocalPort  int
	RemotePort int
	Status     ForwardStatus
	Err        string
	StartedAt  time.Time
}

type ForwardEvent struct {
	SessionID string
	TargetID  string
	Status    ForwardStatus
	Err       string
}
