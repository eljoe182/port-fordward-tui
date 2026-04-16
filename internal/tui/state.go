package tui

type CatalogItem struct {
	Type               string
	Namespace          string
	Name               string
	ID                 string
	Label              string
	RemotePort         int
	PreferredLocalPort int
	Favorite           bool
	Available          bool
}

type SelectedItem struct {
	TargetID   string
	Label      string
	LocalPort  int
	RemotePort int
}

type ForwardStatus string

const (
	StatusStarting ForwardStatus = "starting"
	StatusRunning  ForwardStatus = "running"
	StatusStopped  ForwardStatus = "stopped"
	StatusFailed   ForwardStatus = "failed"
)

type RunningItem struct {
	TargetID   string
	SessionID  string
	Context    string
	Namespace  string
	Type       string
	Label      string
	LocalPort  int
	RemotePort int
	Status     ForwardStatus
	Err        string
}

type RuntimeEvent struct {
	TargetID string
	Status   ForwardStatus
	Err      string
}
