package tui

type CatalogItem struct {
	Context            string
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
	Context    string
	Namespace  string
	Type       string
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

type forwardRef struct {
	Context   string
	Namespace string
	TargetID  string
}

func (item CatalogItem) ref() forwardRef {
	return forwardRef{Context: item.Context, Namespace: item.Namespace, TargetID: item.ID}
}

func (item SelectedItem) ref() forwardRef {
	return forwardRef{Context: item.Context, Namespace: item.Namespace, TargetID: item.TargetID}
}

func (item RunningItem) ref() forwardRef {
	return forwardRef{Context: item.Context, Namespace: item.Namespace, TargetID: item.TargetID}
}
